package bloxid

import (
	"crypto/rand"
)

const (
	DefaultUniqueIDDecodedCharSize = 40                                     // must be even number and multiple of 40 bits
	DefaultUniqueIDEncodedCharSize = DefaultUniqueIDDecodedCharSize * 5 / 8 // must be divisible by 8 without remainder
	DefaultUniqueIDByteSize        = DefaultUniqueIDDecodedCharSize / 2     // must be divisible by 2 without remainder
	DefaultEntropySize             = DefaultUniqueIDDecodedCharSize         // deprecated but kept for back compat
)

func randDefault() []byte {
	return randBytes(DefaultUniqueIDByteSize)
}

func randBytes(size int) []byte {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
