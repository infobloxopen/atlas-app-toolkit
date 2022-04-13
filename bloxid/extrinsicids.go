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
		o.schemer = &schemerExtrinsicID{
			extrinsicID: eid,
			scheme:      IDSchemeExtrinsic,
		}
	}
}

type schemerExtrinsicID struct {
	extrinsicID string
	scheme      string
}

var _ Schemer = (*schemerRandomEncodedID)(nil)

func (sch *schemerExtrinsicID) FromEntityID(opts *V0Options) (scheme string, decoded string, encoded string, err error) {
	decoded, err = getExtrinsicID(sch.extrinsicID)
	if err != nil {
		return
	}

	encoded = encodeLowerAlphaNumeric(extrinsicIDPrefix, decoded)
	scheme = IDSchemeExtrinsic
	return
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
