package query

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
)

// DecodePageToken decodes page token from the user's request.
// Return error if provided token is malformed or contains ivalid values,
// otherwise return offset, limit.
func DecodePageToken(ptoken string) (offset, limit int32, err error) {
	errC := errors.InitContainer()
	data, err := base64.StdEncoding.DecodeString(ptoken)
	if err != nil {
		return 0, 0, errC.New(codes.InvalidArgument, "Invalid page token %q.", err)
	}
	vals := strings.SplitN(string(data), ":", 2)
	if len(vals) != 2 {
		return 0, 0, errC.New(codes.InvalidArgument, "Malformed page token.")
	}

	o, err := strconv.Atoi(vals[0])
	if err != nil {
		errC.Set("page_token", codes.InvalidArgument, "invalid offset value %q.", vals[0])
		errC.WithField("offset", "Invalid offset value. The valid value is an unsigned integer.")
	}

	l, err := strconv.Atoi(vals[1])
	if err != nil {
		errC.Set("page_token", codes.InvalidArgument, "invalid limit value %q.", vals[1])
		errC.WithField("limit", "Invalid limit value. The valid value is an unsigned integer.")
	}

	limit = int32(l)
	offset = int32(o)

	if err := errC.IfSet(codes.InvalidArgument, "Page token validation failed."); err != nil {
		return 0, 0, errC
	}

	return offset, limit, nil
}

// EncodePageToken encodes offset and limit to a string in application specific
// format (offset:limit) in base64 encoding.
func EncodePageToken(offset, limit int32) string {
	data := fmt.Sprintf("%d:%d", offset, limit)
	return base64.StdEncoding.EncodeToString([]byte(data))
}
