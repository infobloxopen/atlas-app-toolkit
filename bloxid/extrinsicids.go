package bloxid

import (
	"errors"
	"regexp"
)

const (
	extrinsicIDPrefix = "EXTR" //the prefix needs to be uppercase
)

var (
	extrinsicIDPrefixBytes = []byte(extrinsicIDPrefix)
)

var (
	ErrEmptyExtrinsicID   = errors.New("empty extrinsic id")
	ErrInvalidExtrinsicID = errors.New("invalid extrinsic id")

	extrinsicIDRegex = regexp.MustCompile(`^[0-9A-Za-z_-]+$`)
)

func validateGetExtrinsicID(id string) error {
	if len(id) < 1 {
		return ErrEmptyExtrinsicID
	}

	if !extrinsicIDRegex.MatchString(id) {
		return ErrInvalidExtrinsicID
	}

	return nil
}

func getExtrinsicID(id string) (string, error) {
	if err := validateGetExtrinsicID(id); err != nil {
		return "", err
	}

	return id, nil
}
