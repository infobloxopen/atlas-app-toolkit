package internal

import "strings"

const (
	delimiter = "/"
)

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

	return strings.Join(l, delimiter)
}

func ParseString(id string) (aname, rtype, rid string) {
	v := strings.SplitN(strings.Trim(id, delimiter), delimiter, 3)
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
