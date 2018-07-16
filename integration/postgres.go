package integration

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	errConnectionTimeout = testDatabaseError{
		message: "database connection timed out",
	}
)

type testDatabaseError struct{ message string }

func (e testDatabaseError) Error() string {
	return fmt.Sprintf("testing database error: %s", e.message)
}

type testPostgresDB struct {
	host            string
	port            int
	dbName          string
	dbUser          string
	dbPassword      string
	dbVersion       string
	migrateFunction func(sql.DB) error
	timeout         time.Duration
}

// NewTestPostgresDB returns a test postgres database that
// the functional options that have been provided by the caller
func NewTestPostgresDB(opts ...option) (testPostgresDB, error) {
	port, err := GetOpenPortInRange(35000, portRangeMax)
	if err != nil {
		return testPostgresDB{}, err
	}
	db := testPostgresDB{
		port:       port,
		dbName:     "test-postgres-db",
		dbUser:     "postgres",
		dbPassword: "postgres",
		dbVersion:  "latest",
		timeout:    time.Second * 10,
	}
	for _, opt := range opts {
		opt(&db)
	}
	return db, nil
}

// Reset drops all the tables in a test database and regenerates them by
// running migrations. If a migration function has not been specified, then the
// tables are dropped but not regenerated
func (db testPostgresDB) Reset() error {
	dbSQL, err := sql.Open("postgres", db.GetDSN())
	if err != nil {
		return err
	}
	defer dbSQL.Close()
	resetQuery := "DROP SCHEMA public CASCADE;" +
		"CREATE SCHEMA public;" +
		"GRANT ALL ON SCHEMA public TO postgres;" +
		"GRANT ALL ON SCHEMA public TO public;"
	// drop all the tables in the test database
	if _, err := dbSQL.Exec(resetQuery); err != nil {
		return err
	}
	// run migrations if a migration function has exists
	if db.migrateFunction != nil {
		if err := db.migrateFunction(*dbSQL); err != nil {
			return err
		}
	}
	return nil
}

// RunAsDockerContainer spins-up a Postgres database server as a Docker
// container. The test Postgres database will run inside this Docker container.
func (db testPostgresDB) RunAsDockerContainer(containerName string) (func() error, error) {
	cleanup, err := RunContainer(
		// define the postgres image version
		fmt.Sprintf("postgres:%s", db.dbVersion),
		// define the arguments to docker
		[]string{
			fmt.Sprintf("--name=%s", containerName),
			fmt.Sprintf("--publish=%d:5432", db.port),
			fmt.Sprintf("--env=POSTGRES_DB=%s", db.dbName),
			fmt.Sprintf("--env=POSTGRES_PASSWORD=%s", db.dbPassword),
			fmt.Sprintf("--env=POSTGRES_USER=%s", db.dbUser),
			"--detach",
			"--rm",
		},
		// define the runtime arguments to postgres
		[]string{},
	)
	if err != nil {
		return nil, err
	}
	if err := db.checkConnection(); err != nil {
		cleanup()
		return nil, err
	}
	return cleanup, nil
}

func (db testPostgresDB) checkConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	errStream := make(chan error)
	go func() {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				driver, _ := sql.Open("postgres", db.GetDSN())
				if err := driver.Ping(); err == nil {
					errStream <- nil
					return
				}
			case <-ctx.Done():
				errStream <- errConnectionTimeout
			}
		}
	}()
	if err := <-errStream; err != nil {
		return err
	}
	return nil
}

// GetDSN returns the database connection string for the test Postgres database
func (db testPostgresDB) GetDSN() string {
	return fmt.Sprintf(
		"host=localhost port=%d user=%s password=%s sslmode=disable dbname=%s",
		db.port, db.dbUser, db.dbPassword, db.dbName,
	)
}

type option func(*testPostgresDB)

// WithName is used to specify the name of the test Postgres database. By default
// the database name is "test-postgres-db"
func WithName(name string) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.dbName = name
	}
}

// WithPort is used to specify the port of the test Postgres database. By
// default, the test database will find the first open port in the 35000+ range
func WithPort(port int) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.port = port
	}
}

// WithUser is used to specify the name of the Postgres user that owns the
// test database
func WithUser(user string) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.dbUser = user
	}
}

// WithPassword is used to specify the password of the test Postgres database
func WithPassword(password string) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.dbPassword = password
	}
}

// WithVersion is used to specify the version of the test Postgres database. By
// default the version is "latest"
func WithVersion(version string) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.dbVersion = version
	}
}

// WithMigrateFunction is used to rebuild the test Postgres database on a
// per-test basis. Whenever the database is reset with the Reset() function, the
// migrate function will rebuild the tables.
func WithMigrateFunction(migrateFunction func(sql.DB) error) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.migrateFunction = migrateFunction
	}
}

// WithTimeout is used to specify a connection timeout to the database
func WithTimeout(timeout time.Duration) func(*testPostgresDB) {
	return func(db *testPostgresDB) {
		db.timeout = timeout
	}
}
