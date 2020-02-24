package health

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
)

type ChecksHandler struct {
	lock sync.RWMutex

	livenessPath   string
	livenessChecks map[string]Check

	readinessPath   string
	readinessChecks map[string]Check

	logger *logrus.Logger
}

// Checker ...
type Checker interface {
	AddLiveness(name string, check Check)
	AddReadiness(name string, check Check)
	Handler() http.Handler
	RegisterHandler(mux *http.ServeMux)
}

// NewChecksHandler accepts two strings: health and ready paths.
// These paths will be used for liveness and readiness checks.
func NewChecksHandler(healthPath, readyPath string) Checker {
	return newChecksHandler(healthPath, readyPath)
}

func newChecksHandler(healthPath, readyPath string) *ChecksHandler {
	if healthPath[0] != '/' {
		healthPath = "/" + healthPath
	}
	if readyPath[0] != '/' {
		readyPath = "/" + readyPath
	}
	ch := &ChecksHandler{
		livenessPath:    healthPath,
		livenessChecks:  map[string]Check{},
		readinessPath:   readyPath,
		readinessChecks: map[string]Check{},
		logger:          nil,
	}
	return ch
}

func NewChecksHandlerWithOptions(healthPath, readyPath string, options ...func(*ChecksHandler)) *ChecksHandler {
	ch := newChecksHandler(healthPath, readyPath)

	for _, option := range options {
		option(ch)
	}

	return ch
}

func WithLogger(logger *logrus.Logger) func(*ChecksHandler) {
	return func(c *ChecksHandler) {
		c.logger = logger
	}
}

func (ch *ChecksHandler) AddLiveness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.logger != nil {
		ch.logger.WithFields(logrus.Fields{
			"liveness_path":  ch.livenessPath,
			"readiness_path": ch.readinessPath,
			"name":           name,
		}).Warn("adding liveness check")
	}

	ch.livenessChecks[name] = check
}

func (ch *ChecksHandler) AddReadiness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.logger != nil {
		ch.logger.WithFields(logrus.Fields{
			"liveness_path":  ch.livenessPath,
			"readiness_path": ch.readinessPath,
			"name":           name,
		}).Warn("adding liveness check")
	}

	ch.readinessChecks[name] = check
}

// Handler returns a new http.Handler for the given health checker
func (ch *ChecksHandler) Handler() http.Handler {
	if ch.logger != nil {
		ch.logger.WithFields(logrus.Fields{
			"liveness_path":  ch.livenessPath,
			"readiness_path": ch.readinessPath,
		}).Warn("creating new ServeMux for health checker")
	}
	mux := http.NewServeMux()
	ch.registerMux(mux)
	return mux
}

// RegisterHandler registers the given health and readiness patterns onto the given http.ServeMux
func (ch *ChecksHandler) RegisterHandler(mux *http.ServeMux) {
	if ch.logger != nil {
		ch.logger.WithFields(logrus.Fields{
			"liveness_path":  ch.livenessPath,
			"readiness_path": ch.readinessPath,
		}).Warn("registering ServeMux for health checker")
	}
	ch.registerMux(mux)
}

func (ch *ChecksHandler) registerMux(mux *http.ServeMux) {
	if ch.logger != nil {
		ch.logger.WithFields(logrus.Fields{
			"liveness_path":  ch.livenessPath,
			"readiness_path": ch.readinessPath,
		}).Warn("registering endpoints for health checker")
	}
	mux.HandleFunc(ch.readinessPath, ch.readyEndpoint)
	mux.HandleFunc(ch.livenessPath, ch.healthEndpoint)
}

func (ch *ChecksHandler) healthEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.livenessChecks)
}

func (ch *ChecksHandler) readyEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.readinessChecks)
}

func checkLogger(logger *logrus.Entry) (*logrus.Entry, bool) {
	if logger == nil {
		return logger, false
	} else if logger.Logger.Out == ioutil.Discard {
		return logger, false
	}
	return logger, true
}

func (ch *ChecksHandler) handle(rw http.ResponseWriter, r *http.Request, checksSets ...map[string]Check) {
	logger := ch.logger
	ctxLogger, ok := checkLogger(ctxlogrus.Extract(r.Context()))
	if ok {
		logger = ctxLogger.Logger
	}

	if r.Method != http.MethodGet {
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"liveness_path":  ch.livenessPath,
				"readiness_path": ch.readinessPath,
				"url":            r.URL.RawPath,
			}).Warn("received non-GET request")
		}
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	errors := map[string]error{}
	status := http.StatusOK
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	for _, checks := range checksSets {
		for name, check := range checks {
			if check == nil {
				continue
			}
			if err := check(); err != nil {
				if logger != nil {
					logger.WithFields(logrus.Fields{
						"liveness_path":  ch.livenessPath,
						"readiness_path": ch.readinessPath,
						"url":            r.URL.RawPath,
					}).WithError(err).Error("health check returned error")
				}
				status = http.StatusServiceUnavailable
				errors[name] = err
			}
		}
	}
	rw.WriteHeader(status)

	return

	// Uncomment to write errors and get non-empty response
	// rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	// if status == http.StatusOK {
	// 	rw.Write([]byte("{}\n"))
	// } else {
	// 	encoder := json.NewEncoder(rw)
	// 	encoder.SetIndent("", "    ")
	// 	encoder.Encode(errors)
	// }
}
