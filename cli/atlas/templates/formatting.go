package templates

import (
	"fmt"
	"strings"
	"unicode"
)

var (
	errEmptyServiceName   = formatError{"empty service name"}
	errInvalidFirstRune   = formatError{"leading non-letter in service name"}
	errInvalidProjectRoot = formatError{"project must be initialized inside $GOPATH/src directory or below"}
	errMissingGOPATH      = formatError{"$GOPATH environment variable not set"}
)

// ServiceName takes a string and formats it into a valid gRPC service name
func ServiceName(str string) (string, error) {
	if len(str) < 1 {
		return str, errEmptyServiceName
	}
	if first := rune(str[0]); !unicode.IsLetter(first) {
		return str, errInvalidFirstRune
	}
	// split string on one or more non-alphanumeric runes
	fields := strings.FieldsFunc(str, isSpecial)
	for i := range fields {
		fields[i] = strings.Title(fields[i])
	}
	return strings.Join(fields, ""), nil
}

// ServerURL takes a string and forms a valid URL string
func ServerURL(str string) (string, error) {
	if len(str) < 1 {
		return str, errEmptyServiceName
	}
	// split string on one or more non-alphanumeric runes
	fields := strings.FieldsFunc(str, isSpecial)
	url := strings.Join(fields, "-")
	return strings.ToLower(url), nil
}

// ProjectRoot determines the root directory of an application. The project
// root is considered to be anything in $GOPATH/src or below
func ProjectRoot(gopath, workdir string) (string, error) {
	if gopath == "" {
		return "", errMissingGOPATH
	}
	if len(workdir) < len(gopath) {
		return "", errInvalidProjectRoot
	}
	if workdir[:len(gopath)] != gopath {
		return "", errInvalidProjectRoot
	}
	projectpath := workdir[len(gopath):]
	dirs := strings.Split(projectpath, "/")
	if len(dirs) < 2 {
		return "", errInvalidProjectRoot
	}
	return strings.Join(dirs[2:], "/"), nil
}

// isSpecial checks if rune is non-alphanumeric
func isSpecial(r rune) bool {
	return (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z')
}

type formatError struct {
	msg string
}

func (e formatError) Error() string {
	return fmt.Sprintf("formatting error: %s", e.msg)
}
