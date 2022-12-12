package tracing

import (
	"runtime"
)

// ThisFunction returns calling function name
func ThisFunction() string {
	pc := make([]uintptr, 32)
	runtime.Callers(2, pc)
	return runtime.FuncForPC(pc[0]).Name()
}

// Trace returns calling function file name, line number and name
func Trace() (file string, line int, name string) {
	pc := make([]uintptr, 32)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line = f.FileLine(pc[0])
	return file, line, f.Name()
}
