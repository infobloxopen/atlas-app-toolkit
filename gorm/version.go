package gorm

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	jgorm "github.com/jinzhu/gorm"
)

// MigrationVersionValidator has a function for checking the database version
type MigrationVersionValidator interface {
	ValidVersion(int64) error
}

type rangeVersion struct {
	lower, upper int64
}
type singleVersion struct {
	target int64
}

// ValidVersion returns an appropriate error if the version is outside the expected range
func (v *rangeVersion) ValidVersion(version int64) error {
	if version < v.lower {
		return fmt.Errorf("Database at version %d, lower than requirement of %d", version, v.lower)
	}
	if version > v.upper {
		return fmt.Errorf("Database at version %d, higher than requirement of %d", version, v.upper)
	}
	return nil
}

// ValidVersion returns an appropriate error if the version is not equal to the target version
func (v *singleVersion) ValidVersion(version int64) error {
	if version != v.target {
		return fmt.Errorf("Database at version %d, not equal to requirement of %d", version, v.target)
	}
	return nil
}

// VersionRange returns a MigrationVersionValidator with a given lower and upper bound
func VersionRange(lower, upper int64) MigrationVersionValidator {
	return &rangeVersion{
		lower: lower,
		upper: upper,
	}
}

// VersionExactly returns a MigrationVersionValidator with a specific target version
func VersionExactly(version int64) MigrationVersionValidator {
	return &singleVersion{
		target: version,
	}
}

// MaxVersionFrom returns a MigrationVersionValidator with a target based on the
// highest numbered migration file detected in the given directory
func MaxVersionFrom(path string) (MigrationVersionValidator, error) {
	version := &singleVersion{}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return version, err
	}
	for _, f := range files {
		if f.IsDir() || !strings.Contains(f.Name(), ".") {
			continue
		}
		parts := strings.Split(f.Name(), "_")
		ext := f.Name()[strings.LastIndex(f.Name(), "."):]
		if ext != ".sql" {
			continue
		}
		if len(parts) < 2 {
			return version, fmt.Errorf("Filename %q does not match migration file naming requirements ##_name.[up/down].sql", f.Name())
		}
		if nVer, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
			return version, err
		} else if nVer > version.target {
			version.target = nVer
		}
	}
	return version, nil
}

// VerifyMigrationVersion checks the schema_migrations table of the db passed
// against the ValidVersion function of the given validator, returning an error
// for an invalid version or a dirty database
func VerifyMigrationVersion(db *jgorm.DB, v MigrationVersionValidator) error {
	var version int64
	var dirty bool
	row := db.DB().QueryRow(`SELECT * FROM schema_migrations`)
	if err := row.Scan(&version, &dirty); err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("Database at version %d, but is dirty", version)
	}
	if err := v.ValidVersion(version); err != nil {
		return err
	}
	return nil
}
