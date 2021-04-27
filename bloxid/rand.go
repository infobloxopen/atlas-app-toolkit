package bloxid

import (
	"crypto/rand"
)

const (
	DefaultEntropySize = 40
)

func randDefault() []byte {
	return randBytes(DefaultEntropySize)
}

func randBytes(size int) []byte {
	b := make([]byte, size/2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
