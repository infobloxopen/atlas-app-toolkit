package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if len(os.Getenv("e2e")) == 0 {
		log.Print("skipping end-to-end tests")
		return
	}

	if err := os.RemoveAll("test"); err != nil {
		log.Fatalf("failed to delete test folder: %v", err)
	}
	log.Print("installing atlas cli")
	if out, err := exec.Command("go", "install").CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to install atlas cli: %v", err)
	}
	log.Print("running init-app")
	if out, err := exec.Command("atlas", "init-app", "-name=test", "-gateway").CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to run atlas init-app: %v", err)
	}
	defer func() {
		log.Print("cleaning up bootstrapped files")
		if err := os.RemoveAll("test"); err != nil {
			log.Fatalf("failed to delete test folder: %v", err)
		}
	}()

	packages := []string{"server", "gateway"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, pkg := range packages {
		output := fmt.Sprintf("./test/bin/%s", pkg)
		log.Printf("building %s", pkg)
		build := exec.Command("go", "build", "-o", output, fmt.Sprintf("./test/cmd/%s", pkg))
		if out, err := build.CombinedOutput(); err != nil {
			log.Print(string(out))
			log.Fatalf("failed to build %s: %v", pkg, err)
		}
		log.Printf("runnning %s", pkg)
		if err := exec.CommandContext(ctx, output).Start(); err != nil {
			log.Fatalf("failed to start server %s: %v", pkg, err)
		}
	}
	log.Print("wait for servers to load up")
	time.Sleep(time.Second)

	m.Run()
}

func TestGetVersion(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/test/v1/version")
	if err != nil {
		t.Errorf("expected get/version to succeed, but got error: %v", err)
	} else if resp.StatusCode != 200 {
		t.Errorf("expected response to be status 200, but got %d: %v", resp.StatusCode, resp)
	}
}
