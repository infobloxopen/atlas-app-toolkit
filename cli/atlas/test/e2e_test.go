package test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("e2e") == "true" {
		os.Exit(m.Run())
	}
}

func TestGeneratedCodeBuilds(t *testing.T) {
	t.Log("running e2e test!")
	t.Fail()
}
