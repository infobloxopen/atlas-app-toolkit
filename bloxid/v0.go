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

	IDSchemeRandom = "random"
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

const (
	V0Delimiter   = "."
	bloxidTypeLen = 4
)

type EncodeDecodeOpts func(o *V0Options)

// NewV0 parse a string into a typed guid, return an error
// if the string fails validation.
func NewV0(bloxid string, fnOpts ...EncodeDecodeOpts) (*V0, error) {
	opts := generateV0Options(fnOpts...)

	if len(strings.TrimSpace(bloxid)) == 0 {
		return generateV0(opts)
	}

	return parseV0(bloxid, opts.hashidSalt)
}

func generateV0Options(fnOpts ...EncodeDecodeOpts) *V0Options {
	var opts *V0Options = new(V0Options)

	for _, fn := range fnOpts {
		fn(opts)
	}

	return opts
}

func parseV0(bloxid, salt string) (*V0, error) {
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
	case bytes.HasPrefix(decoded, hashIDPrefixBytes):
		var hashed string
		hashed = strings.TrimSpace(string(decoded[bloxidTypeLen:]))
		v0.hashIDInt64, err = getInt64FromHashID(hashed, salt)
		if err != nil {
			return nil, err
		}
		v0.decoded = fmt.Sprintf("%v", v0.hashIDInt64)
		v0.scheme = IDSchemeHashID
	case bytes.HasPrefix(decoded, extrinsicIDPrefixBytes):
		v0.decoded = strings.TrimSpace(string(decoded[bloxidTypeLen:]))
		v0.scheme = IDSchemeExtrinsic
	default:
		if len(v0.encoded) < DefaultUniqueIDEncodedCharSize {
			return nil, ErrInvalidUniqueIDLen
		}

		v0.decoded = hex.EncodeToString(decoded)
		v0.scheme = IDSchemeRandom
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
	entityDomain string
	entityType   string
	realm        string
	decoded      string
	encoded      string
	hashIDInt64  int64
	scheme       string
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

// Version of the string
func (v *V0) Version() string {
	if v == nil {
		return VersionUnknown.String()
	}
	return v.version.String()
}

// Domain implements ID.domain
func (v *V0) Domain() string {
	if v == nil {
		return ""
	}
	return v.entityDomain
}

// Type implements ID.entityType
func (v *V0) Type() string {
	if v == nil {
		return ""
	}
	return v.entityType
}

// Realm implements ID.realm
func (v *V0) Realm() string {
	if v == nil {
		return ""
	}
	return v.realm
}

// DecodedID implements ID.decoded
func (v *V0) DecodedID() string {
	if v == nil {
		return ""
	}
	return v.decoded
}

// EncodedID implements ID.encoded
func (v *V0) EncodedID() string {
	if v == nil {
		return ""
	}
	return v.encoded
}

// HashIDInt64 implements ID.hashIDInt64
func (v *V0) HashIDInt64() int64 {
	if v == nil || v.scheme != IDSchemeHashID {
		return -1
	}
	return v.hashIDInt64
}

// Scheme of the id
func (v *V0) Scheme() string {
	if v == nil {
		return ""
	}
	return v.scheme
}

// V0Options required options to create a typed guid
type V0Options struct {
	entityDomain string
	entityType   string
	realm        string
	extrinsicID  string
	hashIDInt64  int64
	hashidSalt   string
	scheme       string
}

type GenerateV0Opts func(o *V0Options)

func WithEntityDomain(domain string) func(o *V0Options) {
	return func(o *V0Options) {
		o.entityDomain = domain
	}
}

func WithEntityType(eType string) func(o *V0Options) {
	return func(o *V0Options) {
		o.entityType = eType
	}
}

func WithRealm(realm string) func(o *V0Options) {
	return func(o *V0Options) {
		o.realm = realm
	}
}

func generateV0(opts *V0Options, fnOpts ...GenerateV0Opts) (*V0, error) {
	for _, fn := range fnOpts {
		fn(opts)
	}

	encoded, decoded, err := uniqueID(opts)
	if err != nil {
		return nil, err
	}

	return &V0{
		version:      Version0,
		realm:        opts.realm,
		decoded:      decoded,
		encoded:      encoded,
		entityDomain: opts.entityDomain,
		entityType:   opts.entityType,
		hashIDInt64:  opts.hashIDInt64,
		scheme:       opts.scheme,
	}, nil
}

func uniqueID(opts *V0Options) (encoded, decoded string, err error) {

	switch opts.scheme {
	case IDSchemeHashID:
		var hashed string

		if opts.hashIDInt64 < 0 {
			err = ErrInvalidID
			return
		}

		decoded = fmt.Sprintf("%v", opts.hashIDInt64)
		hashed, err = getHashID(opts.hashIDInt64, opts.hashidSalt)
		if err != nil {
			return
		}

		encoded = encodeLowerAlphaNumeric(hashIDPrefix, hashed)

	case IDSchemeExtrinsic:

		decoded, err = getExtrinsicID(opts.extrinsicID)
		if err != nil {
			return
		}

		encoded = encodeLowerAlphaNumeric(extrinsicIDPrefix, decoded)

	default:
		rndm := randDefault()
		decoded = hex.EncodeToString(rndm)
		encoded = strings.ToLower(base32.StdEncoding.EncodeToString(rndm))
		opts.scheme = IDSchemeRandom
	}

	return
}

func encodeLowerAlphaNumeric(idPrefix, decoded string) string {

	const rfc4648NoPaddingChars = 5
	rem := rfc4648NoPaddingChars - ((len(decoded) + len(idPrefix)) % rfc4648NoPaddingChars)
	pad := strings.Repeat(" ", rem)
	padded := idPrefix + decoded + pad
	return strings.ToLower(base32.StdEncoding.EncodeToString([]byte(padded)))
}
