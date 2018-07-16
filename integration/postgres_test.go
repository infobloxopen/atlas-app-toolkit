package integration

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// buildDB creates a new test Postgres database and halts the test if anything
// fails during this process
func buildDB(t *testing.T, opts ...option) testPostgresDB {
	db, err := NewTestPostgresDB(opts...)
	if err != nil {
		t.Fatalf("unable to create test postgres database")
	}
	return db
}

func TestGetDSN(t *testing.T) {
	var tests = []struct {
		name     string
		opts     []option
		expected string
	}{
		{
			name: "vanilla test database",
			opts: []option{
				// test will fail if default port is taken, so the port is specified
				WithPort(37000),
			},
			expected: "host=localhost port=37000 user=postgres " +
				"password=postgres sslmode=disable dbname=test-postgres-db",
		},
		{
			name: "test with all options",
			opts: []option{
				WithName("some-test-db"),
				WithPort(37000),
				WithUser("test-user"),
				WithPassword("secret"),
				WithVersion("9.5.13"),
			},
			expected: "host=localhost port=37000 user=test-user " +
				"password=secret sslmode=disable dbname=some-test-db",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, err := NewTestPostgresDB(test.opts...)
			if err != nil {
				t.Fatalf("unable to build test database")
			}
			if dsn := db.GetDSN(); dsn != test.expected {
				t.Errorf("unexpected database connection string: have %s; want %s",
					dsn, test.expected,
				)
			}
		})
	}
}

func TestCheckConnection(t *testing.T) {
	db, err := NewTestPostgresDB(
		WithTimeout(time.Second * 5),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	db.host = "some-fake-hostname"
	if err := db.checkConnection(); err == nil {
		t.Errorf("expected to get non-nil error")
	}
	db.host = "localhost"
	rm, err := db.RunAsDockerContainer("test-postgres-container")
	if err != nil {
		t.Fatalf("unable to start the database container: %v", err)
	}
	defer rm()
}

type testTable struct {
	gorm.Model
	Test string
}

func TestReset(t *testing.T) {
	db, err := NewTestPostgresDB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.Reset(); err == nil {
		t.Errorf("expected to receive an error when resetting database")
	}
	rm, err := db.RunAsDockerContainer("test-postgres-container")
	if err != nil {
		t.Fatalf("unable to start the database container: %v", err)
	}
	defer rm()
	db.migrateFunction = func(*sql.DB) error {
		return errors.New("intentional testing error")
	}
	if err := db.Reset(); err == nil {
		t.Errorf("expected to receive an error when migrating database")
	}
	db.migrateFunction = func(*sql.DB) error {
		orm, err := gorm.Open("postgres", db.GetDSN())
		if err != nil {
			t.Errorf("unable to connect to database: %v", err)
		}
		orm.AutoMigrate(&testTable{})
		return nil
	}
	if err := db.Reset(); err != nil {
		t.Errorf("unexpected error when resetting the database: %v", err)
	}
}
