package bloxid

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	VersionUnknown Version = iota
	Version0       Version = iota
)

type Version uint8

func (v Version) String() string {
	name, ok := version_name[v]
	if !ok {
		return version_name[VersionUnknown]
	}
	return name
}

var version_name = map[Version]string{
	VersionUnknown: "unknown",
	Version0:       "blox0",
}

var name_version = map[string]Version{
	"unknown": VersionUnknown,
	"blox0":   Version0,
}

var (
	ErrInvalidVersion     error = errors.New("invalid bloxid version")
	ErrInvalidEntityType  error = errors.New("entity type must be non-empty")
	ErrInvalidUniqueIDLen error = errors.New("unique ID did not meet minimum length requirements")

	ErrIDEmpty error = errors.New("empty bloxid")
	ErrV0Parts error = errors.New("invalid number of parts found")
)

// NewV0 parse a string into a typed guid, return an error
// if the string fails validation.
func NewV0(bloxid string) (*V0, error) {
	return parseV0(bloxid)
}

const V0Delimiter = "."

func parseV0(bloxid string) (*V0, error) {
	if len(bloxid) == 0 {
		return nil, ErrIDEmpty
	}

	if err := validateV0(bloxid); err != nil {
		return nil, err
	}

	parts := strings.Split(bloxid, V0Delimiter)
	v0 := &V0{
		version:    name_version[parts[0]],
		entityType: parts[1],
		region:     parts[2],
		encoded:    parts[3],
	}

	decodedTall := strings.ToUpper(parts[3])
	decoded, err := base32.StdEncoding.DecodeString(decodedTall)
	if err != nil {
		return nil, fmt.Errorf("unable to decode id: %s", err)
	}
	v0.decoded = hex.EncodeToString(decoded)

	return v0, nil
}

func validateV0(bloxid string) error {
	parts := strings.Split(bloxid, V0Delimiter)
	if len(parts) != 4 {
		return ErrV0Parts
	}

	if ver := parts[0]; ver != Version0.String() {
		return ErrInvalidVersion
	}

	if entityType := parts[1]; len(entityType) == 0 {
		return ErrInvalidEntityType
	}

	if len(parts[3]) < DefaultUniqueIDEncodedCharSize {
		return ErrInvalidUniqueIDLen
	}

	return nil
}

var _ ID = &V0{}

// V0 represents a typed guid
type V0 struct {
	version      Version
	region       string
	customSuffix string
	decoded      string
	encoded      string
	entityType   string
	shortID      string
}

// Serialize the typed guid as a string
func (v *V0) String() string {
	s := []string{
		v.Version(),
		v.entityType,
		v.region,
		v.encoded,
	}
	return strings.Join(s, V0Delimiter)
}

// Region implements ID.Region
func (v *V0) Region() string {
	if v == nil {
		return ""
	}
	return v.region
}

// ShortID implements ID.ShortID
func (v *V0) ShortID() string {
	if v == nil {
		return ""
	}
	return v.shortID
}

// Type implements ID.Type
func (v *V0) Type() string {
	if v == nil {
		return ""
	}
	return v.entityType
}

// Version of the string
func (v *V0) Version() string {
	if v == nil {
		return VersionUnknown.String()
	}
	return v.version.String()
}

// V0Options required options to create a typed guid
type V0Options struct {
	Region     string
	EntityType string
	shortid    string
}

type GenerateV0Opts func(o *V0Options)

func WithShortID(shortid string) func(o *V0Options) {
	return func(o *V0Options) {
		o.shortid = shortid
	}
}

func GenerateV0(opts *V0Options, fnOpts ...GenerateV0Opts) (*V0, error) {

	for _, fn := range fnOpts {
		fn(opts)
	}

	encoded, decoded := uniqueID(opts)

	return &V0{
		version:    Version0,
		region:     opts.Region,
		decoded:    decoded,
		encoded:    encoded,
		entityType: opts.EntityType,
	}, nil
}

func uniqueID(opts *V0Options) (encoded string, decoded string) {
	rndm := randDefault()
	decoded = hex.EncodeToString(rndm)
	encoded = strings.ToLower(base32.StdEncoding.EncodeToString(rndm))
	return
}
