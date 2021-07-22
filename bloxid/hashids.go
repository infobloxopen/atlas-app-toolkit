package bloxid

import (
	"errors"

	hashids "github.com/speps/go-hashids/v2"
)

const (
	IDSchemeHashID = "hashid"

	hashIDAllowedChar = "0123456789abcdef"

	//the prefix needs to be uppercase so there are no collisions with chars in hashIDAllowedChar
	hashIDPrefix = "HIDZ"
)

var (
	ErrInvalidSalt = errors.New("invalid salt")
	ErrInvalidID   = errors.New("invalid id")

	maxHashIDLen = DefaultUniqueIDDecodedCharSize - len(hashIDPrefix)

	hashIDPrefixBytes = []byte(hashIDPrefix)
)

func newHashID(salt string) (*hashids.HashID, error) {
	hid := hashids.HashIDData{
		Alphabet:  hashIDAllowedChar,
		MinLength: maxHashIDLen,
		Salt:      salt,
	}

	return hashids.NewWithData(&hid)
}

func validateGetHashID(id int64, salt string) error {

	if len(salt) < 1 {
		return ErrInvalidSalt
	}

	if id < 0 {
		return ErrInvalidID
	}

	return nil
}

func getHashID(id int64, salt string) (string, error) {
	if err := validateGetHashID(id, salt); err != nil {
		return "", err
	}

	h, err := newHashID(salt)
	if err != nil {
		return "", err
	}

	eID, err := h.EncodeInt64([]int64{id})
	if err != nil {
		return "", err
	}

	return eID, err
}

func validateGetint64FromHashID(id, salt string) error {

	if len(salt) < 1 {
		return ErrInvalidSalt
	}

	if len(id) != maxHashIDLen {
		return ErrInvalidID
	}

	return nil
}

func getInt64FromHashID(id, salt string) (int64, error) {
	if err := validateGetint64FromHashID(id, salt); err != nil {
		return -1, err
	}

	h, err := newHashID(salt)
	if err != nil {
		return -1, err
	}

	dID, err := h.DecodeInt64WithError(id)
	if err != nil {
		return -1, err
	}

	return dID[0], err
}

func WithHashIDInt64(id int64) func(o *V0Options) {
	return func(o *V0Options) {
		o.hashIDInt64 = id
		o.scheme = IDSchemeHashID
	}
}

func WithHashIDSalt(salt string) func(o *V0Options) {
	return func(o *V0Options) {
		o.hashidSalt = salt
	}
}
