package gorm

import (
	"errors"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

func TestVersionFromPath(t *testing.T) {
	v, err := MaxVersionFrom("testdata/fake_migrations")
	if err != nil {
		t.Errorf("Unexpected error %q", err.Error())
	}
	expected := &singleVersion{
		target: 3,
	}
	if !reflect.DeepEqual(v, expected) {
		t.Errorf("Expected %+v but got %+v", expected, v)
	}
}

func TestVersionInDB(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hasV     int64
		hasDirty bool
		checkV   MigrationVersionValidator
		expErr   error
	}{
		{
			name:     "version exact correct",
			hasV:     3,
			checkV:   VersionExactly(3),
			hasDirty: false,
			expErr:   nil,
		},
		{
			name:     "version exactly wrong",
			hasV:     5,
			checkV:   VersionExactly(3),
			hasDirty: false,
			expErr:   errors.New("Database at version 5, not equal to requirement of 3"),
		},
		{
			name:     "version range correct",
			hasV:     3,
			checkV:   VersionRange(1, 4),
			hasDirty: false,
			expErr:   nil,
		},
		{
			name:     "version too low",
			hasV:     3,
			checkV:   VersionRange(1, 2),
			hasDirty: false,
			expErr:   errors.New("Database at version 3, higher than requirement of 2"),
		},
		{
			name:     "version too high",
			hasV:     3,
			checkV:   VersionRange(4, 5),
			hasDirty: false,
			expErr:   errors.New("Database at version 3, lower than requirement of 4"),
		},
		{
			name:     "version dirty",
			hasV:     3,
			checkV:   VersionRange(1, 5),
			hasDirty: true,
			expErr:   errors.New("Database at version 3, but is dirty"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock - %s", err)
			}
			vrows := sqlmock.NewRows([]string{"version", "dirty"})
			vrows.AddRow(tc.hasV, tc.hasDirty)
			mock.ExpectQuery(`SELECT \* FROM schema_migrations`).WillReturnRows(vrows)

			gdb, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm db - %s", err)
			}

			err = VerifyMigrationVersion(gdb, tc.checkV)
			if tc.expErr != nil {
				if !reflect.DeepEqual(err, tc.expErr) {
					t.Errorf("Was supposed to return error (%s) but returned (%s)", tc.expErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("failed to verify mocked version - %s", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("failed to query properly - %s", err)
			}
		})
	}
}
