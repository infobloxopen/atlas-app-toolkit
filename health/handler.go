package health

import (
	"net/http"
	"sync"
)

type checksHandler struct {
	lock            sync.RWMutex
	mux             *http.ServeMux
	livenessChecks  map[string]Check
	readinessChecks map[string]Check
}

// Checker ...
type Checker interface {
	AddLiveness(name string, check Check)
	AddReadiness(name string, check Check)
	Handler() http.Handler
}

// NewChecksHandler accepts two strings: health and ready paths.
// These paths will be used for liveness and readiness checks.
func NewChecksHandler(healthPath, readyPath string) Checker {
	ch := &checksHandler{
		livenessChecks:  map[string]Check{},
		readinessChecks: map[string]Check{},
		mux:             &http.ServeMux{},
	}
	if healthPath[0] != '/' {
		healthPath = "/" + healthPath
	}
	if readyPath[0] != '/' {
		readyPath = "/" + readyPath
	}
	ch.mux.Handle(healthPath, http.HandlerFunc(ch.healthEndpoint))
	ch.mux.Handle(readyPath, http.HandlerFunc(ch.readyEndpoint))
	return ch
}

func (ch *checksHandler) AddLiveness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.livenessChecks[name] = check
}

func (ch *checksHandler) AddReadiness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.readinessChecks[name] = check
}

func (ch *checksHandler) Handler() http.Handler {
	return ch.mux
}

func (ch *checksHandler) healthEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.livenessChecks)
}

func (ch *checksHandler) readyEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.readinessChecks, ch.livenessChecks)
}

func (ch *checksHandler) handle(rw http.ResponseWriter, r *http.Request, checksSets ...map[string]Check) {
	if r.Method != http.MethodGet {
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
