package cmode

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	urlPath       = "/cmode"
	valuesUrlPath = urlPath + "/values"
)

type CMode struct {
	logger CModeLogger
	opts   []CModeOpt
	usage  []string
}

type CModeOpt interface {
	Name() string
	Get() string
	ParseAndSet(val string) error
	Description() string
	ValidValues() []string
}

type CModeLogger interface {
	CModeOpt
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

func New(cmLogger CModeLogger, opts ...CModeOpt) CMode {
	cm := CMode{
		opts:   []CModeOpt{},
		usage:  []string{},
		logger: cmLogger,
	}

	cm.AddOption(cmLogger)
	for _, opt := range opts {
		if !isOptNil(opt) {
			cm.opts = append(cm.opts, opt)
		}
	}

	cm.generateUsage()
	return cm
}

func Handler(cm CMode) http.Handler {
	h := mux.NewRouter()
	h.HandleFunc(urlPath, cm.help).Methods("GET")
	h.HandleFunc(valuesUrlPath, cm.get).Methods("GET")
	h.HandleFunc(valuesUrlPath, cm.set).Methods("POST")
	return h
}

func (cm *CMode) AddOption(opt CModeOpt) {
	if !isOptNil(opt) {
		cm.opts = append(cm.opts, opt)
		cm.generateUsage()
	}
}

func (cm *CMode) generateUsage() {
	maxPathLength := len(valuesUrlPath)
	for _, opt := range cm.opts {
		path := fmt.Sprintf("%s?%s=$%s", valuesUrlPath, opt.Name(), strings.ToUpper(opt.Name()))
		if len(path) > maxPathLength {
			maxPathLength = len(path)
		}
	}

	baseString := strings.Repeat(" ", maxPathLength)

	usage := []string{}

	usage = append(usage, "Usage:")
	usage = append(usage, fmt.Sprintf("GET  %s%s -- print usage", urlPath, baseString[6:]))
	usage = append(usage, fmt.Sprintf("GET  %s%s -- get current values", valuesUrlPath, baseString[13:]))

	if len(cm.opts) > 0 {
		for _, opt := range cm.opts {
			path := fmt.Sprintf("%s?%s=$%s", valuesUrlPath, opt.Name(), strings.ToUpper(opt.Name()))

			sb := strings.Builder{}
			sb.WriteString(path)
			sb.WriteString(baseString[len(path):])
			path = sb.String()

			usage = append(usage, fmt.Sprintf("POST %s -- %s", path, opt.Description()))
		}

		usage = append(usage, "")

		for _, opt := range cm.opts {
			v := fmt.Sprintf("valid %s values: [ %s ]", opt.Name(), strings.Join(opt.ValidValues(), ", "))
			usage = append(usage, v)
		}
	}

	cm.usage = usage
}

func (cm *CMode) help(w http.ResponseWriter, _ *http.Request) {
	writeReply(w, http.StatusOK, cm.usage)
}

func (cm *CMode) get(w http.ResponseWriter, _ *http.Request) {
	var reply []string
	for _, opt := range cm.opts {
		reply = append(reply, fmt.Sprintf("%s: %s", opt.Name(), opt.Get()))
	}
	writeReply(w, http.StatusOK, reply)
}

func (cm *CMode) set(w http.ResponseWriter, r *http.Request) {
	var reply []string
	empty := true

	for _, opt := range cm.opts {
		optVal := r.URL.Query().Get(opt.Name())
		if optVal != "" {
			empty = false
			err := opt.ParseAndSet(optVal)
			if err != nil {
				replyText := fmt.Sprintf("invalid %s value: %s", opt.Name(), optVal)
				reply = append(reply, replyText)
				reply = append(reply, cm.usage...)
				writeReply(w, http.StatusBadRequest, reply)
				if !cm.isLoggerNil() {
					cm.logger.Errorf(replyText)
				}
				return
			}
			if !cm.isLoggerNil() {
				cm.logger.Infof("%s is set to %s", opt.Name(), optVal)
			}
		}
	}

	if empty {
		writeReply(w, http.StatusBadRequest, cm.usage)
		return
	}

	writeReply(w, http.StatusOK, reply)
}

func (cm *CMode) isLoggerNil() bool {
	switch v := cm.logger.(type) {
	case CModeLogger:
		return reflect.ValueOf(v).IsNil()
	case nil:
		return true
	default:
		panic("You're trying to pass a var that doesn't implement 'CMode' interface")
	}
}

func writeReply(w http.ResponseWriter, status int, reply []string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(strings.Join(reply, "\n")))
	if err != nil {
		logrus.Errorf("error writing reply: %s", err)
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		logrus.Errorf("error writing reply: %s", err)
	}
}

func isOptNil(opt CModeOpt) bool {
	switch v := opt.(type) {
	case CModeOpt:
		return reflect.ValueOf(v).IsNil()
	case nil:
		return true
	default:
		panic("You're trying to pass a var that doesn't implement 'CModeOpt' interface")
	}
}
