package health

import (
	"net/http"
	"sync"
)

type checksContextHandler struct {
	lock sync.RWMutex

	livenessPath   string
	livenessChecks map[string]CheckContext

	readinessPath   string
	readinessChecks map[string]CheckContext

	// if true first found error will fail the check stage
	failFast bool
}

// CheckerContext ...
type CheckerContext interface {
	AddLiveness(name string, check CheckContext)
	AddReadiness(name string, check CheckContext)
	Handler() http.Handler
	RegisterHandler(mux *http.ServeMux)
	SetFailFast(failFast bool)
	GetFailFast() bool
}

// NewChecksContextHandler accepts two strings: health and ready paths.
// These paths will be used for liveness and readiness checks.
func NewChecksContextHandler(healthPath, readyPath string) CheckerContext {
	if healthPath[0] != '/' {
		healthPath = "/" + healthPath
	}
	if readyPath[0] != '/' {
		readyPath = "/" + readyPath
	}
	ch := &checksContextHandler{
		livenessPath:    healthPath,
		livenessChecks:  map[string]CheckContext{},
		readinessPath:   readyPath,
		readinessChecks: map[string]CheckContext{},
	}

	return ch
}

// SetFailFast sets failFast flag for failing on the first error found
func (ch *checksContextHandler) SetFailFast(failFast bool) {
	ch.failFast = failFast
}

func (ch *checksContextHandler) GetFailFast() bool {
	return ch.failFast
}

func (ch *checksContextHandler) AddLiveness(name string, check CheckContext) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.livenessChecks[name] = check
}

func (ch *checksContextHandler) AddReadiness(name string, check CheckContext) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.readinessChecks[name] = check
}

// Handler returns a new http.Handler for the given health checker
func (ch *checksContextHandler) Handler() http.Handler {
	mux := http.NewServeMux()
	ch.registerMux(mux)
	return mux
}

// RegisterHandler registers the given health and readiness patterns onto the given http.ServeMux
func (ch *checksContextHandler) RegisterHandler(mux *http.ServeMux) {
	ch.registerMux(mux)
}

func (ch *checksContextHandler) registerMux(mux *http.ServeMux) {
	mux.HandleFunc(ch.readinessPath, ch.readyEndpoint)
	mux.HandleFunc(ch.livenessPath, ch.healthEndpoint)
}

func (ch *checksContextHandler) healthEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.livenessChecks)
}

func (ch *checksContextHandler) readyEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.readinessChecks)
}

func (ch *checksContextHandler) handle(rw http.ResponseWriter, r *http.Request, checksSets ...map[string]CheckContext) {
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
			if err := check(r.Context()); err != nil {
				status = http.StatusServiceUnavailable
				errors[name] = err
				if ch.failFast {
					rw.WriteHeader(status)
					return
				}
			}
		}
	}
	rw.WriteHeader(status)

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
