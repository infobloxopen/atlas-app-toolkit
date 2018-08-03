package resource

// String method calls proto.CompactTextString which returns "<nil>" string
// in case if passed proto.Message is nil.
// The "<nil>" string is common representation of nil value in Go code. See fmt package.
const pbnil = "<nil>"

// Valid reports whether id valid ot not.
// The id is valid if it is neither nil nor empty string.
func Valid(id *Identifier) (ok bool) {
	// comparison with pbnil is mostly paranoid check
	// should never happen
	if id == nil || id.String() == "" || id.String() == pbnil {
		return
	}
	return true
}
