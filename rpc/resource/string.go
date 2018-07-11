package resource

import "strings"

const (
	// Delimiter of Resource Reference according to Atlas Reference format
	Delimiter = "/"
)

// BuildString builds string id according to Atlas Reference format:
//  <application_name>/<resource_type>/<resource_id>
func BuildString(aname, rtype, rid string) string {
	var l []string

	if v := strings.TrimSpace(aname); v != "" {
		l = append(l, v)
	}
	if v := strings.TrimSpace(rtype); v != "" {
		l = append(l, v)
	}
	if v := strings.TrimSpace(rid); v != "" {
		l = append(l, v)
	}

	return strings.Join(l, Delimiter)
}

// ParseString parses id according to Atlas Reference format:
//	<application_name>/<resource_type>/<resource_id>
// All leading and trailing Delimiter are removed.
// The resource_id is parsed first, then resource type and last application name.
// The id "/a/b/c/" will be converted to "a/b/c" and returned as (a, b, c).
// The id "b/c/" will be converted to "b/c" and returned as ("", b, c).
func ParseString(id string) (aname, rtype, rid string) {
	v := strings.SplitN(strings.Trim(id, Delimiter), Delimiter, 3)
	switch len(v) {
	case 1:
		rid = v[0]
	case 2:
		rtype, rid = v[0], v[1]
	case 3:
		aname, rtype, rid = v[0], v[1], v[2]
	}
	return
}

func (m Identifier) MarshalText() (text []byte, err error) {
	text = []byte(BuildString(m.GetApplicationName(), m.GetResourceType(), m.GetResourceId()))
	return
}
