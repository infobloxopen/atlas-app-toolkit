package bloxid

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	IDSchemeRandom = "random"

	RandomEncodedIDSize = 32
)

var (
	ErrEmptyRandomEncodedID           = errors.New("empty random scheme encoded id")
	ErrInvalidSizeRandomEncodedID     = fmt.Errorf("random scheme encoded id must be %d chars", RandomEncodedIDSize)
	ErrInvalidAlphabetRandomEncodedID = errors.New("invalid random scheme encoded id")
	ErrInvalidRandomEncodedID         = errors.New("invalid random scheme encoded id")

	rRandomEncodedIDRegex = regexp.MustCompile(`^[0-9a-z]+$`)
)

// WithRandomEncodedID supplies the unique id portion of a previously generated random scheme bloxid
func WithRandomEncodedID(eid string) func(o *V0Options) {
	return func(o *V0Options) {
		o.schemer = &schemerRandomEncodedID{
			encodedID: eid,
			scheme:    IDSchemeRandom,
		}
	}
}

type schemerRandomEncodedID struct {
	encodedID string
	scheme    string
}

var _ Schemer = (*schemerRandomEncodedID)(nil)

func (sch *schemerRandomEncodedID) FromEntityID(opts *V0Options) (scheme string, decoded string, encoded string, err error) {
	decoded, err = getDecodedIDFromRandomEncodedID(sch.encodedID)
	if err != nil {
		return
	}
	encoded = sch.encodedID
	scheme = IDSchemeRandom
	return
}

func validateGetRandomEncodedID(id string) error {
	if ln := len(id); ln < 1 {
		return ErrEmptyRandomEncodedID
	} else if ln != RandomEncodedIDSize {
		return ErrInvalidSizeRandomEncodedID
	}

	if !rRandomEncodedIDRegex.MatchString(id) {
		return ErrInvalidAlphabetRandomEncodedID
	}

	return nil
}

func getDecodedIDFromRandomEncodedID(id string) (string, error) {
	if err := validateGetRandomEncodedID(id); err != nil {
		return "", err
	}

	val, err := base32.StdEncoding.DecodeString(strings.ToUpper(id))
	if err != nil {
		return "", ErrInvalidRandomEncodedID
	}
	decoded := hex.EncodeToString(val)

	return decoded, nil
}
