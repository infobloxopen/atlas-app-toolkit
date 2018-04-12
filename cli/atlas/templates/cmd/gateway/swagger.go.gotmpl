package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// SwaggerHandler returns an HTTP handler that serves the swagger spec
func SwaggerHandler(rw http.ResponseWriter, req *http.Request) {
	p := strings.TrimPrefix(req.URL.Path, "/swagger/")
	p = filepath.Join(SwaggerDir, p)
	p += ".swagger.json"
	log.Printf("serving %s", p)
	http.ServeFile(rw, req, p)
}
