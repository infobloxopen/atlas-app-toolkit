package query

// GoString implements fmt.GoStringer interface
// return string representation of searching in next form:
func (s *Searching) GoString() string {
	return s.Query
}

// ParseSearching parses raw string that represent search criteria into a Searching
// data structure.
func ParseSearching(s string) *Searching {
	var searching Searching
	searching.Query = s
	return &searching
}
