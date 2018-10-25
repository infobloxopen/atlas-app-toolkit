# Integration

This package contains helper functions that support integration testing of Go services and applications. The main library utilities are listed below.
- Starting and stopping Docker containers inside Go tests
- Creating JSON Web Tokens for testing gRPC and REST requests
- Launching a Postgres database to tests against
- Building and executing Go binaries

## Examples
### Start and Stop Docker Containers
Testing your application or service in isolation might be impossible. Your project may require a database or supporting services in order to function reliably. 

Enter Docker containers.

You can prop-up Docker containers to substitute backend databases or services. This package is intended to make that process easier with helper functions. To start out, you might want to familiarize yourself with Go's built-in [`TestMain`](http://cs-guy.com/blog/2015/01/test-main) function, but here's the basic gist.

> It is sometimes necessary for a test program to do extra setup or teardown before or after testing. It is also sometimes necessary for a test to control which code runs on the main thread. `TestMain` runs in the main goroutine and can do whatever setup and teardown is necessary.

Awesome. Example time!


```go
import (
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

// TestMain does pre-test set up
func TestMain(m *testing.M) {
  // RunContainer takes a docker image, docker run arguments, and 
  // runtime-specific arguments
  stop, err := integration.RunContainer(
    "redis:latest",
    []string{
      "--publish=6380:6380",
      "--env=REDIS_PASSWORD=password",
      "--rm",
    },
    []string{
      "maxmemory 2mb",
    },
  )
  if err != nil {
    log.Fatal("unable to start test redis container")
  }
  // stop and remove container after testing
  defer stop() 
  m.Run()
}
```

### Launching a Postgres Database

Your application or service might be backed by a Postgres database. This library can help configure, launch, and reset a Postgres database between tests.

To start out, here's how you can use the Postgres options to configure your database.

```go
import (
	"database/sql"
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
	_ "github.com/lib/pq"
)

var (
	myTestDatabase integration.TestPostgresDB
)

func TestMain(m *testing.M) {
	// myMigrateFunc describes how to build the database schema from scratch
	myMigrateFunc := func(db *sql.DB) error {
		// migration.Run(db) isn't a real function, but let's pretend it
		// creates some tables in a database
		return migration.Run(db)
	}
	config, err := integration.NewTestPostgresDB(
		integration.WithName("my_database_name"),
		// passing a migrate up function will allow the database schema to be
		// destroyed and re-built, effectively causing the database to reset
		integration.WithMigrateUpFunction(myMigrateFunc),
	)
	if err != nil {
		log.Fatal("unable to build postgres config")
	}
	myTestDatabase = config
	stop, err := myTestDatabase.RunAsDockerContainer()
	if err != nil {
		log.Fatal("unable to start test database")
	}
	defer stop()
	m.Run()
}
```

The example above will configure and launch the database. The code below shows how to reset the database between tests.

```go
import (
  "testing"
)

func TestMyEndpoint(t *testing.T) {
  // reset the database before running tests. you would want to do this if any
  // other tests create or modify entries in the database
  if err := myTestDatabase.Reset(); err != nil{
    t.Fatalf("unable to reset database schema: %v", err)
  }
  ... 
}
```

### Building and Running Go Binaries

If you want to test your Go application or service, you'll need to build it first. The `integration` package provides helpers that enable you to build your Go binary and run it locally.

Alternatively, you can build your application or service's Docker image, then run the Docker image by following the [earlier examples](#Start and Stop Docker Containers).

#### Building the Binary

```go
import (
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMain(m *testing.M) {
  remove, err := integration.BuildGoSource("./path/to/my/go/package", "binaryName")
  if err != nil {
    log.Fatalf("unable to build go binary: %v", err)
  }
  // this will delete the binary after the tests run
  defer remove()
  m.Run()
}
```
#### Running the Binary

```go
import (
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMain(m *testing.M) {
	stop, err := integration.RunBinary(
		"./path/to/my/go/package/binaryName",
		// provide as many command-line arguments as you like
		"-debug=true",
	)
	if err != nil {
		log.Fatalf("unable to run go binary: %v", err)
	}
	// this will stop the running process
	defer stop()
	m.Run()
}

```
### Finding Open Ports

To help avoid port conflicts, the `integration` package provides a simple helper that finds an port on the testing machine.


```go
import (
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMain(m *testing.M) {
	port, err := integration.GetOpenPort()
	if err != nil {
		log.Fatalf("unable to find open port: %v", err)
	}
}

```

You can also specify a port range.

```go
import (
	"log"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMain(m *testing.M) {
	port, err := integration.GetOpenPortInRange(6000, 8000)
	if err != nil {
		log.Fatalf("unable to find open port: %v", err)
	}
}
```

### Creating JSON Web Tokens

If you plan to run test requests against your application or service, you might need to provide a JWT for authentication purposes. This isn't terribly tricky, but it's nice to have some helpers that spare you from reinventing the wheel.

#### Using the Standard Token

If you just need a token, but don't particularly care what it contains, then you might want to use the standard token. The term _standard_ just means the token has the minimum required JWT claims that are needed to authenticate.

```go
import (
	"net/http"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMyEndpoint(t *testing.T) {
	token, err := integration.StandardTestJWT()
	if err != nil {
		t.Fatalf("unable to generate test token: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "/endpoint", nil)
	if err != nil {
		t.Fatalf("unable to generate test request: %v", err)
	}
	req.Header.Set("Authorization: Bearer %s", token)
	...
}
```

### Creating Default Test Requests

You might want to create REST requests or gRPC requests that use the standard JWT. Rather than write code that packs the JWT into the HTTP request header, or the gRPC request context, the integration library has utilities to do this for you.

Here's how you would be a test HTTP request.

```go
import (
	"net/http"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMyEndpoint(t *testing.T) {
	client := http.Client{}
	req, err := integration.MakeStandardRequest(
		http.MethodGet, "/endpoint", map[string]string{
			"message": "hello world",
		},
	)
	if err != nil {
		t.Fatalf("unable to build test http request: %v", err)
	}
	res, err := client.Do(req)
	...
}
```

And the same for gRPC requests.

```go
import (
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/integration"
)

func TestMyGRPCEndpoint(t *testing.T) {
	ctx, err := integration.StandardTestingContext()
	if err != nil {
		t.Fatalf("unable to build test grpc context: %v", err)
	}
	gRPCResponseMessage, err := gRPCClient.MyGRPCEndpoint(ctx, gRPCRequestMessage)
	if err != nil {
		t.Fatalf("unable to send grpc request: %v", err)
	}
	...
}
```

