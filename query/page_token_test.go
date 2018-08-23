package query

import (
	"testing"
)

func TestDecodePageToken(t *testing.T) {
	tcases := []struct {
		PageToken     string
		Offset        int32
		Limit         int32
		ExpectedError string
	}{
		{
			PageToken:     "asd", //invalid
			ExpectedError: "Invalid page token \"illegal base64 data at input byte 0\".",
		},
		{
			PageToken:     "MTI=", //12
			ExpectedError: "Malformed page token.",
		},
		{
			PageToken:     "YTpi", //a:b
			ExpectedError: "Page token validation failed.",
		},
		{
			PageToken:     "MTI6Yg==", //12:b
			ExpectedError: "Page token validation failed.",
		},
		{
			PageToken:     "YToxMg==", //a:12
			ExpectedError: "Page token validation failed.",
		},
		{
			PageToken: "MTI6MzQ=", //12:34
			Limit:     34,
			Offset:    12,
		},
		{
			PageToken: "MzQ6MTI=", //34:12
			Limit:     12,
			Offset:    34,
		},
	}

	for n, tc := range tcases {
		offset, limit, err := DecodePageToken(tc.PageToken)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %q, expected %q", n, err, tc.ExpectedError)
		}
		if limit != tc.Limit {
			t.Fatalf("tc %d: invalid limit %d, expected %d", n, limit, tc.Limit)
		}
		if offset != tc.Offset {
			t.Fatalf("tc %d: invalid offset %d, expected %d", n, offset, tc.Offset)
		}
	}
}

func TestEncodePageToken(t *testing.T) {
	tcases := []struct {
		Offset    int32
		Limit     int32
		PageToken string
	}{
		{
			Limit:     34,
			Offset:    12,
			PageToken: "MTI6MzQ=",
		},
		{
			Limit:     12,
			Offset:    34,
			PageToken: "MzQ6MTI=",
		},
	}

	for n, tc := range tcases {
		ptoken := EncodePageToken(tc.Offset, tc.Limit)
		if ptoken != tc.PageToken {
			t.Fatalf("tc %d: invalid page token %q, expected %q", n, ptoken, tc.PageToken)
		}
	}
}
