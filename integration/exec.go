package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunBinary runs a target binary with a set of arguments provided by the user
func RunBinary(binPath string, args ...string) (func(), error) {
	abs, err := filepath.Abs(binPath)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, abs, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cancel, nil
}

// BuildGoSource builds a target Go package and gives the resulting binary
// some user-defined name. The function returned by BuildGoSource will remove
// the binary that got created.
func BuildGoSource(packagePath, output string) (func() error, error) {
	cmdBuild := exec.Command("go", "build", "-o", output, packagePath)
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("unable to build package: %v (%s)", err, string(out))
	}
	return func() error {
		return os.Remove(output)
	}, nil
}

// RunContainer launches a detached docker container on the host machine. It
// takes an image name, a list of "docker run" arguments, and a list of
// arguments that get passed to the container runtime
func RunContainer(image string, dockerArgs, runtimeArgs []string) (func() error, error) {
	// build a list of args to pass to cmd.Exec
	execDocker := append([]string{"docker", "run", "-d"}, dockerArgs...)
	execRuntime := append([]string{image}, runtimeArgs...)
	execDocker = append(execDocker, execRuntime...)
	// launch the container
	out, err := exec.Command(execDocker[0], execDocker[1:]...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), string(out))
	}
	// get the uuid from the container after it starts. this gets used to kill
	// the container
	split := strings.Split(string(out), "\n")
	if len(split) == 0 {
		return nil, errors.New("unable to get container uuid")
	}
	uuid := split[0]
	return func() error {
		return exec.Command("docker", "kill", uuid).Run()
	}, nil
}
