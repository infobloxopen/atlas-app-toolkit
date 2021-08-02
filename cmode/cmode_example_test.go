package cmode

import (
	"net/http"

	"github.com/infobloxopen/atlas-app-toolkit/cmode/logger"
	"github.com/sirupsen/logrus"
)

func ExampleCMode() {
	appLogger := logrus.New()

	cmLogger := logger.New(appLogger)
	cm := New(appLogger, &cmLogger)

	http.Handle("/", Handler(cm))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		appLogger.Fatalf("Server fatal error - %s", err)
	}
}
