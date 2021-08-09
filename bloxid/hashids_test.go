package bloxid

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashIDtoInt(t *testing.T) {
	var tests = []struct {
		name    string
		hashID  string
		int64ID int64
		salt    string
		err     error
	}{
		{
			name:    "Valid input",
			hashID:  "blox0.infra.host.us-com-1.jbeuiwrsmq3tkmzwmuzwcojsmrqwemrtgy3tqzbvhbsdizjvhe2dkn3cgzrdizlb",
			int64ID: 1,
			salt:    "test",
			err:     nil,
		},
		{
			name:    "Different salt",
			hashID:  "blox0.infra.host.us-com-1.jbeuiwrsmq3tkmzwmuzwcojsmrqwemrtgy3tqzbvhbsdizjvhe2dkn3cgzrdizlb",
			int64ID: -1,
			salt:    "testi1",
			err:     errors.New("mismatch between encode and decode: 2d7536e3a92dab23678d58d4e59457b6b4ea start ed4b2a9764524ed6958d58237ba3eadb7695 re-encoded. result: [4]"),
		},

		{
			name:    "Invalid prefix HIDA with correct int value",
			hashID:  "blox0.infra.host.us-com-1.jbeuiqjsmq3tkmzwmuzwcojsmrqwemrtgy3tqzbvhbsdizjvhe2dkn3cgzrdizlb",
			int64ID: -1,
			salt:    "test",
			err:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v0_1, err := NewV0(test.hashID, WithHashIDSalt(test.salt))

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.int64ID, v0_1.HashIDInt64())
		})
	}
}

func TestHashIDInttoInt(t *testing.T) {
	var tests = []struct {
		name       string
		int64ID    int64
		entityType string
		domainType string
		realm      string
		salt       string
		err        error
	}{
		{
			name:       "Valid input",
			int64ID:    1,
			entityType: "hostapp",
			domainType: "infra",
			realm:      "us-com-1",
			salt:       "test",
			err:        nil,
		},
		{
			name:       "Negative number",
			int64ID:    -1,
			entityType: "hostapp",
			domainType: "infra",
			realm:      "us-com-1",
			salt:       "test",
			err:        ErrInvalidID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v0, err := NewV0("",
				WithEntityDomain(test.domainType),
				WithEntityType(test.entityType),
				WithRealm(test.realm),
				WithHashIDInt64(test.int64ID),
				WithHashIDSalt(test.salt))

			assert.Equal(t, test.err, err)

			if err == nil {
				v0_1, err := NewV0(v0.String(), WithHashIDSalt(test.salt))

				assert.Equal(t, test.err, err)
				assert.Equal(t, test.int64ID, v0_1.HashIDInt64())
				decoded, err := strconv.ParseInt(v0_1.DecodedID(), 10, 64)
				assert.NoError(t, err)
				assert.Equal(t, test.int64ID, decoded)
				assert.Equal(t, v0, v0_1)
			}
		})
	}
}

func TestGetHashID(t *testing.T) {
	var tests = []struct {
		name    string
		int64ID int64
		salt    string
		hashID  string
		err     error
	}{
		{
			name:    "Valid input",
			int64ID: 1,
			salt:    "test",
			hashID:  "2d7536e3a92dab23678d58d4e59457b6b4ea",
			err:     nil,
		},
		{
			name:    "zero number",
			int64ID: 0,
			salt:    "test",
			hashID:  "e735d27d4a5e57d3648658a92be26b93b46a",
			err:     nil,
		},
		{
			name:    "negative number",
			int64ID: -1,
			salt:    "test1",
			hashID:  "",
			err:     ErrInvalidID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hashID, err := getHashID(test.int64ID, test.salt)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.hashID, hashID)
		})
	}
}

func TestGetIntFromHashID(t *testing.T) {
	var tests = []struct {
		name    string
		int64ID int64
		salt    string
		hashID  string
		err     error
	}{
		{
			name:    "Valid input",
			int64ID: 1,
			salt:    "test",
			hashID:  "2d7536e3a92dab23678d58d4e59457b6b4ea",
			err:     nil,
		},
		{
			name:    "negative number",
			int64ID: -1,
			salt:    "test1",
			hashID:  "",
			err:     ErrInvalidID,
		},
		{
			name:    "empty hash",
			int64ID: -1,
			salt:    "test1",
			hashID:  "",
			err:     ErrInvalidID,
		},
		{
			name:    "zero value",
			int64ID: 0,
			salt:    "test",
			hashID:  "e735d27d4a5e57d3648658a92be26b93b46a",
			err:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			int64ID, err := getInt64FromHashID(test.hashID, test.salt)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.int64ID, int64ID)
		})
	}
}

func TestGenerateNewV0(t *testing.T) {
	var testmap = []struct {
		realm        string
		entityDomain string
		entityType   string
		hashIntID    int64
		expected     string
		err          error
	}{
		// ensure `=` is not part of id when encoded
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    1,
			expected:     "blox0.infra.host.us-com-1.ivmfiurreaqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    12,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    123,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgizsaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    1234,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztiiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    12345,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinja",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    123456,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjweaqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    1234567,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg4qcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    12345678,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg44caiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			hashIntID:    123456789,
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg44dsiba",
		},
	}

	for index, tm := range testmap {
		index++
		v0, err := NewV0("",
			WithEntityDomain(tm.entityDomain),
			WithEntityType(tm.entityType),
			WithRealm(tm.realm),
			WithHashIDInt64(tm.hashIntID),
			WithHashIDSalt("test"),
		)
		if err != tm.err {
			//			t.Logf("test: %#v", tm)
			t.Errorf("index: %d got: %s wanted error: %s", index, err, tm.err)
		}
		if err != nil {
			continue
		}

		if v0 == nil {
			t.Errorf("unexpected nil id")
			continue
		}

		if -1 != strings.Index(v0.String(), "=") {
			t.Errorf("got: %q wanted bloxid without equal char", v0.String())
		}
	}
}

/*
func TestHashIDUniqueness(t *testing.T) {

	var salt string = "test"
	var maxIntVal int64 = math.MaxInt64
	var i int64

	checkUniq := make(map[string]struct{})
	for i = 0; i < maxIntVal; i++ {
		v0, err := NewV0("",
			WithHashIDInt64(i),
			WithHashIDSalt(salt),
		)
		if err != nil {
			t.Errorf("Failed to convert int to hash id. Index: %v\n", i)
			continue
		}

		if _, ok := checkUniq[v0.EncodedID()]; ok {
			fmt.Printf("The value is not unique for index: %d - hashid: %s\n", i, v0.EncodedID())
		} else {
			checkUniq[v0.EncodedID()] = struct{}{}
		}

		if -1 != strings.Index(v0.EncodedID(), "=") {
			fmt.Printf("id has equal sign: index %v hashID %v\n", i, v0.EncodedID())
		}
		if i%1000000 == 0 {
			fmt.Printf("Completed: %v\n", i)
		}
	}
	fmt.Println(i)
}
*/
