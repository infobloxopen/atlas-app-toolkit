package resource

// String method calls proto.CompactTextString which returns "<nil>" string
// in case if passed proto.Message is nil.
// The "<nil>" string is common representation of nil value in Go code. See fmt package.
const pbnil = "<nil>"

// Nil reports whether id is empty identifier or not.
// The id is empty if it is either nil or could be converted to the empty string by its String method.
func Nil(id *Identifier) (ok bool) {
	// comparison with pbnil is mostly paranoid check
	// should never happen
	if id == nil || id.String() == "" || id.String() == pbnil {
		return true
	}
	return
}
