package bloxid

import (
	"errors"
	"regexp"
)

const (
	IDSchemeExtrinsic = "extrinsic"

	extrinsicIDPrefix = "EXTR" //the prefix needs to be uppercase
)

var (
	extrinsicIDPrefixBytes = []byte(extrinsicIDPrefix)

	ErrEmptyExtrinsicID   = errors.New("empty extrinsic id")
	ErrInvalidExtrinsicID = errors.New("invalid extrinsic id")

	extrinsicIDRegex = regexp.MustCompile(`^[0-9A-Za-z_-]+$`)
)

// WithExtrinsicID supplies a locally unique ID that is not randomly generated
func WithExtrinsicID(eid string) func(o *V0Options) {
	return func(o *V0Options) {
		o.extrinsicID = eid
		o.scheme = IDSchemeExtrinsic
	}
}

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
