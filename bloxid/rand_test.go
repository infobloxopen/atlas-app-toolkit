package bloxid

import "testing"

func TestRandBytes(t *testing.T) {
	size := 10
	bs := randBytes(10)
	if size != len(bs) {
		t.Errorf("got: %d wanted: %d", len(bs), size)
	}
}
