//go:generate go-bindata -o template-bindata.go templates/...
package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
)

const (
	// the full set of command names
	COMMAND_INIT_APP = "init-app"

	// the full set of flag names
	FLAG_NAME     = "name"
	FLAG_REGISTRY = "registry"
	FLAG_GATEWAY  = "gateway"
)

var (
	// flagset for initializing the application
	initialize         = flag.NewFlagSet(COMMAND_INIT_APP, flag.ExitOnError)
	initializeName     = initialize.String(FLAG_NAME, "", "the application name (required)")
	initializeRegistry = initialize.String(FLAG_REGISTRY, "", "the Docker registry (optional)")
	initializeGateway  = initialize.Bool(FLAG_GATEWAY, false, "generate project with a gRPC gateway (default false)")
)

func main() {
	commandList := []string{COMMAND_INIT_APP}
	if len(os.Args) < 2 {
		fmt.Printf("Command is required. Please choose one of %v\n", fmt.Sprint(commandList))
		os.Exit(1)
	}
	switch command := os.Args[1]; command {
	case COMMAND_INIT_APP:
		initialize.Parse(os.Args[2:])
		initializeApplication()
	default:
		fmt.Printf("Command \"%s\" is not valid. Please choose one of %v\n", command, fmt.Sprint(commandList))
		os.Exit(1)
	}
}

type Application struct {
	Name        string
	Registry    string
	Root        string
	WithGateway bool
}

func (a Application) GenerateDockerfile() {
	a.generateFile("docker/Dockerfile.server", "templates/docker/Dockerfile.application.gotmpl")
}

func (a Application) GenerateGatewayDockerfile() {
	a.generateFile("docker/Dockerfile.gateway", "templates/docker/Dockerfile.gateway.gotmpl")
}

func (a Application) GenerateReadme() {
	a.generateFile("README.md", "templates/README.md.gotmpl")
}

func (a Application) GenerateGitignore() {
	a.generateFile(".gitignore", "templates/.gitignore.gotmpl")
}

func (a Application) GenerateMakefile() {
	a.generateFile("Makefile", "templates/Makefile.gotmpl")
}

func (a Application) GenerateProto() {
	a.generateFile("proto/service.proto", "templates/proto/service.proto.gotmpl")
}

func (a Application) GenerateServerMain() {
	a.generateFile("cmd/server/main.go", "templates/cmd/server/main.go.gotmpl")
}

func (a Application) GenerateGatewayMain() {
	a.generateFile("cmd/gateway/main.go", "templates/cmd/gateway/main.go.gotmpl")
}

func (a Application) GenerateGatewayHandler() {
	a.generateFile("cmd/gateway/handler.go", "templates/cmd/gateway/handler.go.gotmpl")
}

func (a Application) GenerateGatewaySwagger() {
	a.generateFile("cmd/gateway/swagger.go", "templates/cmd/gateway/swagger.go.gotmpl")
}

func (a Application) GenerateConfig() {
	a.generateFile("cmd/config/config.go", "templates/cmd/config/config.go.gotmpl")
}

func (a Application) GenerateService() {
	a.generateFile("svc/zserver.go", "templates/svc/zserver.go.gotmpl")
}

// generateFile creates a file by rendering a template
func (a Application) generateFile(filename, templatePath string) {
	t := template.New("file").Funcs(template.FuncMap{
		"Title":   strings.Title,
		"Service": ServiceName,
		"URL":     ServerURL,
	})
	bytes, err := Asset(templatePath)
	if err != nil {
		panic(err)
	}
	t, err = t.Parse(string(bytes))
	if err != nil {
		panic(err)
	}
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := t.Execute(file, a); err != nil {
		panic(err)
	}
}

// directories returns a list of all project folders
func (a Application) directories() []string {
	dirnames := []string{
		"cmd/server",
		"cmd/config",
		"pb",
		"svc",
		"proto",
		"docker",
		"deploy",
		"migrations",
	}
	if a.WithGateway {
		dirnames = append(dirnames, fmt.Sprintf("cmd/%s", "gateway"))
	}
	return dirnames
}

// initializeApplication generates brand-new application
func initializeApplication() {
	name := *initializeName
	if *initializeName == "" {
		initialize.PrintDefaults()
		os.Exit(1)
	}
	wd, err := os.Getwd()
	if err != nil {
		printErr(err)
	}
	root, err := ProjectRoot(wd)
	if err != nil {
		printErr(err)
	}
	app := Application{name, *initializeRegistry, root, *initializeGateway}
	// initialize project directories
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		msg := fmt.Sprintf("directory '%v' already exists.", name)
		printErr(errors.New(msg))
	}
	os.Mkdir(name, os.ModePerm)
	os.Chdir(name)
	for _, dir := range app.directories() {
		os.MkdirAll(fmt.Sprintf("./%s", dir), os.ModePerm)
	}
	// initialize project files
	app.GenerateDockerfile()
	app.GenerateReadme()
	app.GenerateGitignore()
	app.GenerateMakefile()
	app.GenerateProto()
	app.GenerateServerMain()
	app.GenerateConfig()
	app.GenerateService()
	if app.WithGateway {
		app.GenerateGatewayDockerfile()
		app.GenerateGatewayMain()
		app.GenerateGatewayHandler()
		app.GenerateGatewaySwagger()
	}
	// run post-initialization commands
	if err := generateProtobuf(); err != nil {
		printErr(err)
	}
	if err := resolveImports(app.directories()); err != nil {
		printErr(err)
	}
	if err := initDep(); err != nil {
		printErr(err)
	}
}

func printErr(err error) {
	fmt.Printf("Unable to initialize application: %s\n", err.Error())
	os.Exit(1)
}

// initDep calls "dep init" to generate .toml files
func initDep() error {
	fmt.Print("Starting dep project... ")
	err := exec.Command("dep", "init").Run()
	if err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// generateProtobuf calls "make protobuf" to render initial .pb files
func generateProtobuf() error {
	fmt.Print("Generating protobuf files... ")
	err := exec.Command("make", "protobuf").Run()
	if err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// resolveImports calls "goimports" to determine Go imports
func resolveImports(dirs []string) error {
	fmt.Print("Resolving imports... ")
	for _, dir := range dirs {
		err := exec.Command("goimports", "-w", dir).Run()
		if err != nil {
			return err
		}
	}
	fmt.Println("done!")
	return nil
}
