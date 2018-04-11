# Atlas CLI
This command-line tool helps developers become productive on Atlas. It aims to provide a better development experience by reducing the initial time and effort it takes to build applications.

## Getting Started
These instructions will help you get the Atlas command-line tool up and running on your machine.

### Prerequisites
Please install the following dependencies before running the Atlas command-line tool.

#### goimports

The `goimports` package resolves missing import paths in Go projects. It is part of the [official Go ecosystem](https://golang.org/pkg/#other) and can be installed with the following command:
```
$ go get golang.org/x/tools/cmd/goimports
```
#### dep

This is a dependency management tool for Go. You can install `dep` with Homebrew:

```sh
$ brew install dep
```
More detailed installation instructions are available on the [GitHub repository](https://github.com/golang/dep).

### Installing
The following steps will install the `atlas` binary to your `$GOBIN` directory.

```sh
$ go get github.com/infobloxopen/atlas-app-toolkit/cli/atlas
```
You're all set! Alternatively, you can clone the repository and install the binary manually.

```sh
$ git clone https://github.com/infobloxopen/atlas-app-toolkit.git
$ cd ngp.api.toolkit/cli/atlas
$ go install
```

## Bootstrap an Application
Rather than build applications completely from scratch, you can leverage the command-line tool to initialize a new project. This will generate the necessary files and folders to get started.

```sh
$ atlas init-app name=my-application
$ cd my-application
```
#### Flags
Here's the full set of flags for the `init-app` command.

| Flag          | Description                                                 | Required      | Default Value |
| ------------- | ----------------------------------------------------------- | ------------- | ------------- |
| `name`        | The name of the new application                             | Yes           | N/A           |
| `gateway`     | Initialize the application with a gRPC gateway              | No            | `false`       |
| `registry`    | The Docker registry where application images are pushed     | No            | `""`          |

You can run `atlas init-app --help` to see these flags and their descriptions on the command-line.

#### Additional Examples


```sh
# generates an application with a grpc gateway 
atlas init-app name=my-application -gateway
```

```sh
# specifies a docker registry
atlas init-app name=my-application -registry=infoblox
```
Images names will vary depending on whether or not a Docker registry has been provided.

```sh
# docker registry was provided
registry-name/image-name:image-version
```

```sh
# docker registry was not provided
image-name:image-version
```

Of course, you may include all the flags in the `init-app` command.

```sh
atlas init-app name=my-application -gateway -registry=infoblox
```