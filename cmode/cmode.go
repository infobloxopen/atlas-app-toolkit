package cmode

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	urlPath       = "/cmode"
	valuesUrlPath = urlPath + "/values"
)

type CMode struct {
	logger Logger
	opts   []CModeOpt
	usage  []string
}

type CModeOpt interface {
	Name() string                 // Should be unique. Used in URL - '/cmode/values?$NAME=OPT_VAL'
	Get() string                  // Will be printed in '/cmode/values'
	ParseAndSet(val string) error // Argument is in ValidValues
	Description() string          // Will be printed in '/cmode'
	ValidValues() []string        // Acceptable values. Used in CMode.usage
}

type Logger interface {
	Errorf(format string, args ...interface{}) // Printing message when option setting is failed
	Infof(format string, args ...interface{})  // Printing message when option is successfully set
}

type nopLogger struct{}

func (nl nopLogger) Errorf(format string, args ...interface{}) {}
func (nl nopLogger) Infof(format string, args ...interface{})  {}

func New(cmLogger Logger, opts ...CModeOpt) CMode {

	if cmLogger == nil {
		cmLogger = nopLogger{}
	}

	cm := CMode{
		opts:   opts,
		usage:  []string{},
		logger: cmLogger,
	}

	cm.generateUsage()
	return cm
}

func (cm *CMode) Handler() http.Handler {
	h := mux.NewRouter()
	h.HandleFunc(urlPath, cm.help).Methods("GET")
	h.HandleFunc(valuesUrlPath, cm.get).Methods("GET")
	h.HandleFunc(valuesUrlPath, cm.set).Methods("POST")
	return h
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
	cm.writeReply(w, http.StatusOK, cm.usage)
}

func (cm *CMode) get(w http.ResponseWriter, _ *http.Request) {
	var reply []string
	for _, opt := range cm.opts {
		reply = append(reply, fmt.Sprintf("%s: %s", opt.Name(), opt.Get()))
	}
	cm.writeReply(w, http.StatusOK, reply)
}

func (cm *CMode) set(w http.ResponseWriter, r *http.Request) {
	var reply []string
	empty := true

	for _, opt := range cm.opts {

		optVal := r.URL.Query().Get(opt.Name())
		if optVal == "" {
			continue
		}

		empty = false

		isInValidValues := false
		for _, vv := range opt.ValidValues() {
			if optVal == vv {
				isInValidValues = true
				break
			}
		}

		if !isInValidValues {
			replyText := fmt.Sprintf("invalid %s value: %s", opt.Name(), optVal)
			reply = append(reply, replyText)
			reply = append(reply, cm.usage...)
			cm.writeReply(w, http.StatusBadRequest, reply)
			cm.logger.Errorf(replyText)
			return
		}

		err := opt.ParseAndSet(optVal)
		if err != nil {
			reply = append(reply, "unexpected server error")
			cm.writeReply(w, http.StatusInternalServerError, reply)
			cm.logger.Errorf("%v", err)
			return
		}

		reply = append(reply, fmt.Sprintf("%s is set to %s", opt.Name(), optVal))

		cm.logger.Infof("%s is set to %s", opt.Name(), optVal)
	}

	if empty {
		cm.writeReply(w, http.StatusBadRequest, cm.usage)
		return
	}

	cm.writeReply(w, http.StatusOK, reply)
}

func (cm *CMode) writeReply(w http.ResponseWriter, status int, reply []string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(strings.Join(reply, "\n")))
	if err != nil {
		cm.logger.Errorf("error writing reply: %s", err)
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		cm.logger.Errorf("error writing reply: %s", err)
	}
}
