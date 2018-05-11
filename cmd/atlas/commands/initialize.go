package commands

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"os"
	"os/exec"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/cmd/atlas/templates"
)

const (
	// the full set of command names
	commandInitApp = "init-app"

	// the full set of flag names
	flagAppName      = "name"
	flagRegistryName = "registry"
	flagWithGateway  = "gateway"
	flagWithDebug    = "debug"
	flagWithDatabase = "db"
)

var (
	// flag set for initializing the application
	initialize         = flag.NewFlagSet(commandInitApp, flag.ExitOnError)
	initializeName     = initialize.String(flagAppName, "", "the application name (required)")
	initializeRegistry = initialize.String(flagRegistryName, "", "the Docker registry (optional)")
	initializeGateway  = initialize.Bool(flagWithGateway, false, "generate project with a gRPC gateway (default false)")
	initializeDebug    = initialize.Bool(flagWithDebug, false, "print debug statements during intialization (default false)")
	initializeDatabase = initialize.Bool(flagWithDatabase, false, "initialize the application with database folders")
)

// bootstrap implements the command interface for project intialization
type bootstrap struct{}

func (b bootstrap) GetName() string { return "init-app" }

func (b bootstrap) GetFlagset() *flag.FlagSet { return initialize }

func (b bootstrap) Run() error {
	if *initializeName == "" {
		return initializationError{
			errors.New("application name is required"),
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return initializationError{err: err}
	}
	root, err := templates.ProjectRoot(build.Default.GOPATH, wd)
	if err != nil {
		return initializationError{err: err}
	}
	app := Application{
		Name:         *initializeName,
		Registry:     *initializeRegistry,
		Root:         root,
		WithGateway:  *initializeGateway,
		WithDatabase: *initializeDatabase,
	}
	if err := app.initialize(); err != nil {
		return initializationError{err: err}
	}
	return nil
}

type initializationError struct{ err error }

func (e initializationError) Error() string {
	return fmt.Sprintf("Unable to initialize application: %s", e.err.Error())
}

// Application models the data that the templates need to render files
type Application struct {
	Name         string
	Registry     string
	Root         string
	WithGateway  bool
	WithDatabase bool
}

// initialize generates brand-new application
func (app Application) initialize() error {
	if _, err := os.Stat(app.Name); !os.IsNotExist(err) {
		msg := fmt.Sprintf("directory '%v' already exists.", app.Name)
		return errors.New(msg)
	}
	if err := os.Mkdir(app.Name, os.ModePerm); err != nil {
		return err
	}
	if err := os.Chdir(app.Name); err != nil {
		return err
	}
	if err := app.initializeDirectories(); err != nil {
		return err
	}
	if err := app.initializeFiles(); err != nil {
		return err
	}
	if err := generateProtobuf(); err != nil {
		return err
	}
	if err := initDep(); err != nil {
		return err
	}
	if err := resolveImports(app.getDirectories()); err != nil {
		return err
	}
	if err := initRepo(); err != nil {
		return err
	}
	return nil
}

// initializeFiles generates each file for a new application
func (app Application) initializeFiles() error {
	fileInitializers := []func(Application) error{
		Application.generateDockerfile,
		Application.generateReadme,
		Application.generateGitignore,
		Application.generateMakefile,
		Application.generateProto,
		Application.generateServerMain,
		Application.generateConfig,
		Application.generateService,
		Application.generateServiceTest,
	}
	if app.WithGateway {
		gatewayInitializers := []func(Application) error{
			Application.generateGatewayDockerfile,
			Application.generateGatewayMain,
			Application.generateGatewayHandler,
			Application.generateGatewaySwagger,
		}
		fileInitializers = append(fileInitializers, gatewayInitializers...)
	}
	for _, initializer := range fileInitializers {
		if err := initializer(app); err != nil {
			return err
		}
	}
	return nil
}

// initializeDirectories generates the directory tree for a new application
func (app Application) initializeDirectories() error {
	dirs := app.getDirectories()
	for _, dir := range dirs {
		if err := os.MkdirAll(fmt.Sprintf("./%s", dir), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// getDirectories returns a list of all project folders
func (app Application) getDirectories() []string {
	dirnames := []string{
		"cmd/server",
		"pkg/pb",
		"pkg/svc",
		"docker",
		"deploy",
	}
	if app.WithGateway {
		dirnames = append(dirnames,
			"cmd/gateway",
		)
	}
	if app.WithDatabase {
		dirnames = append(dirnames,
			"db/migrations",
			"db/fixtures",
		)
	}
	return dirnames
}

// generateFile creates a file by rendering a template
func (app Application) generateFile(filename, templatePath string) error {
	t := template.New("file").Funcs(template.FuncMap{
		"Title":    strings.Title,
		"Service":  templates.ServiceName,
		"URL":      templates.ServerURL,
		"Database": templates.DatabaseName,
	})
	bytes, err := templates.Asset(templatePath)
	if err != nil {
		return err
	}
	t, err = t.Parse(string(bytes))
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := t.Execute(file, app); err != nil {
		return err
	}
	return err
}

func (app Application) generateDockerfile() error {
	return app.generateFile("docker/Dockerfile.server", "templates/docker/Dockerfile.application.gotmpl")
}

func (app Application) generateGatewayDockerfile() error {
	return app.generateFile("docker/Dockerfile.gateway", "templates/docker/Dockerfile.gateway.gotmpl")
}

func (app Application) generateReadme() error {
	return app.generateFile("README.md", "templates/README.md.gotmpl")
}

func (app Application) generateGitignore() error {
	return app.generateFile(".gitignore", "templates/.gitignore.gotmpl")
}

func (app Application) generateMakefile() error {
	return app.generateFile("Makefile", "templates/Makefile.gotmpl")
}

func (app Application) generateProto() error {
	return app.generateFile("pkg/pb/service.proto", "templates/pkg/pb/service.proto.gotmpl")
}

func (app Application) generateServerMain() error {
	return app.generateFile("cmd/server/main.go", "templates/cmd/server/main.go.gotmpl")
}

func (app Application) generateGatewayMain() error {
	return app.generateFile("cmd/gateway/main.go", "templates/cmd/gateway/main.go.gotmpl")
}

func (app Application) generateGatewayHandler() error {
	return app.generateFile("cmd/gateway/handler.go", "templates/cmd/gateway/handler.go.gotmpl")
}

func (app Application) generateGatewaySwagger() error {
	return app.generateFile("cmd/gateway/swagger.go", "templates/cmd/gateway/swagger.go.gotmpl")
}

func (app Application) generateConfig() error {
	return app.generateFile("cmd/config.go", "templates/cmd/config.go.gotmpl")
}

func (app Application) generateService() error {
	return app.generateFile("pkg/svc/zserver.go", "templates/pkg/svc/zserver.go.gotmpl")
}

func (app Application) generateServiceTest() error {
	return app.generateFile("pkg/svc/zserver_test.go", "templates/pkg/svc/zserver_test.go.gotmpl")
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if *initializeDebug {
		cmd.Stderr = os.Stdout
		cmd.Stdout = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// generateProtobuf calls "make protobuf" to render initial .pb files
func generateProtobuf() error {
	fmt.Print("Generating protobuf files... ")
	if err := runCommand("make", "protobuf"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// initDep calls "dep init" to generate .toml files
func initDep() error {
	fmt.Print("Starting dep project... ")
	if err := runCommand("dep", "init"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// resolveImports calls "goimports" to determine Go imports
func resolveImports(dirs []string) error {
	fmt.Print("Resolving imports... ")
	for _, dir := range dirs {
		if err := runCommand("goimports", "-w", dir); err != nil {
			return err
		}
	}
	fmt.Println("done!")
	return nil
}

// initRepo initializes new applications as a git repository
func initRepo() error {
	fmt.Print("Initializing git repo... ")
	if err := runCommand("git", "init"); err != nil {
		return err
	}
	if err := runCommand("git", "add", "*"); err != nil {
		return err
	}
	if err := runCommand("git", "commit", "-m", "Initial commit"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}
