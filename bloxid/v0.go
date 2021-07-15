package bloxid

import (
	"bytes"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	UniqueIDEncodedMinCharSize = 16

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
	ErrInvalidVersion      error = errors.New("invalid bloxid version")
	ErrInvalidEntityDomain error = errors.New("entity domain must be non-empty")
	ErrInvalidEntityType   error = errors.New("entity type must be non-empty")
	ErrInvalidUniqueIDLen  error = errors.New("unique ID did not meet minimum length requirements")

	ErrIDEmpty error = errors.New("empty bloxid")
	ErrV0Parts error = errors.New("invalid number of parts found")
)

// NewV0 parse a string into a typed guid, return an error
// if the string fails validation.
func NewV0(bloxid string) (*V0, error) {
	return parseV0(bloxid)
}

const (
	V0Delimiter   = "."
	bloxidTypeLen = 4
)

func parseV0(bloxid string) (*V0, error) {
	if len(bloxid) == 0 {
		return nil, ErrIDEmpty
	}

	if err := validateV0(bloxid); err != nil {
		return nil, err
	}

	parts := strings.Split(bloxid, V0Delimiter)
	v0 := &V0{
		version:      name_version[parts[0]],
		entityDomain: parts[1],
		entityType:   parts[2],
		realm:        parts[3],
		encoded:      parts[4],
	}

	decodedTall := strings.ToUpper(parts[4])
	decoded, err := base32.StdEncoding.DecodeString(decodedTall)
	if err != nil {
		return nil, fmt.Errorf("unable to decode id: %s", err)
	}

	switch {
	case bytes.HasPrefix(decoded, extrinsicIDPrefixBytes):
		v0.decoded = strings.TrimSpace(string(decoded[bloxidTypeLen:]))
	default:
		if len(v0.encoded) < DefaultUniqueIDEncodedCharSize {
			return nil, ErrInvalidUniqueIDLen
		}

		v0.decoded = hex.EncodeToString(decoded)
	}

	return v0, nil
}

func validateV0(bloxid string) error {
	parts := strings.Split(bloxid, V0Delimiter)
	if len(parts) != 5 {
		return ErrV0Parts
	}

	if ver := parts[0]; ver != Version0.String() {
		return ErrInvalidVersion
	}

	if entityDomain := parts[1]; len(entityDomain) == 0 {
		return ErrInvalidEntityDomain
	}

	if entityType := parts[2]; len(entityType) == 0 {
		return ErrInvalidEntityType
	}

	if len(parts[4]) < UniqueIDEncodedMinCharSize {
		return ErrInvalidUniqueIDLen
	}
	return nil
}

var _ ID = &V0{}

// V0 represents a typed guid
type V0 struct {
	version      Version
	realm        string
	decoded      string
	encoded      string
	entityDomain string
	entityType   string
}

// Serialize the typed guid as a string
func (v *V0) String() string {
	s := []string{
		v.Version(),
		v.entityDomain,
		v.entityType,
		v.realm,
		v.encoded,
	}
	return strings.Join(s, V0Delimiter)
}

// Realm implements ID.Realm
func (v *V0) Realm() string {
	if v == nil {
		return ""
	}
	return v.realm
}

// Domain implements ID.Domain
func (v *V0) Domain() string {
	if v == nil {
		return ""
	}
	return v.entityDomain
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
	Realm        string
	EntityDomain string
	EntityType   string
	extrinsicID  string
}

type GenerateV0Opts func(o *V0Options)

// WithExtrinsicID supplies a locally unique ID that is not randomly generated
func WithExtrinsicID(eid string) func(o *V0Options) {
	return func(o *V0Options) {
		o.extrinsicID = eid
	}
}

func GenerateV0(opts *V0Options, fnOpts ...GenerateV0Opts) (*V0, error) {

	for _, fn := range fnOpts {
		fn(opts)
	}

	encoded, decoded := uniqueID(opts)

	return &V0{
		version:      Version0,
		realm:        opts.Realm,
		decoded:      decoded,
		encoded:      encoded,
		entityDomain: opts.EntityDomain,
		entityType:   opts.EntityType,
	}, nil
}

func uniqueID(opts *V0Options) (encoded string, decoded string) {
	if len(opts.extrinsicID) > 0 {
		var err error
		decoded, err = getExtrinsicID(opts.extrinsicID)
		if err != nil {
			return
		}

		const rfc4648NoPaddingChars = 5
		rem := rfc4648NoPaddingChars - ((len(decoded) + len(extrinsicIDPrefix)) % rfc4648NoPaddingChars)
		pad := strings.Repeat(" ", rem)
		padded := extrinsicIDPrefix + decoded + pad
		encoded = strings.ToLower(base32.StdEncoding.EncodeToString([]byte(padded)))
	} else {
		rndm := randDefault()
		decoded = hex.EncodeToString(rndm)
		encoded = strings.ToLower(base32.StdEncoding.EncodeToString(rndm))
	}
	return
}
