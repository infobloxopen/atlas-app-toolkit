package namer

import (
	"strings"
)

// Namer is an interface for allowing cmdline flags
// to be renamed / scoped for use multiple times in a cmd.
type Namer interface {
	FlagNamer(string) string
	EnvVarNamer(string) string
}

type namer struct {
	groupName string
}

// NewNamer returns a Namer implementation
func NewNamer(groupName string) Namer {
	return &namer{groupName}
}

// FlagNamer prepends the input with groupName + "-" and lower-cased
func (n *namer) FlagNamer(input string) string {
	return n.addGroupName(input, "-", false)
}

// EnvVarNamer prepends the input with groupName + ".", lower-cased, and all "-" are replaced with "."
func (n *namer) EnvVarNamer(input string) string {
	return strings.ReplaceAll(n.FlagNamer(input), "-", ".")
}

func (n *namer) addGroupName(input string, separator string, upperCase bool) string {
	resp := input
	if input != "" {
		resp = n.groupName + separator + resp
	}
	if upperCase {
		return strings.ToUpper(resp)
	} else {
		return strings.ToLower(resp)
	}
}
