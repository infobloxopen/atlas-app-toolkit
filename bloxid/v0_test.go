package bloxid

import (
	"strings"
	"testing"
)

func TestNewV0(t *testing.T) {
	var testmap = []struct {
		input        string
		output       string
		entityDomain string
		entityType   string
		decoded      string
		err          error
	}{
		/*
			{
				"",
				"",
				"",
				"",
				"",
				ErrIDEmpty,
			},
		*/
		{
			input: "bloxv0....",
			err:   ErrInvalidVersion,
		},
		{
			input: "blox0....",
			err:   ErrInvalidEntityDomain,
		},
		{
			input: "blox0.infra...",
			err:   ErrInvalidEntityType,
		},
		{
			input: "blox0.infra.host..",
			err:   ErrInvalidUniqueIDLen,
		},
		{
			"blox0.infra.host..zdud52youveke5sovyoc66cjxw3l55jc",
			"blox0.infra.host..zdud52youveke5sovyoc66cjxw3l55jc",
			"infra",
			"host",
			"c8e83eeb0ea548a2764eae1c2f7849bdb6bef522",
			nil,
		},
	}

	for index, tm := range testmap {
		v0, err := NewV0(tm.input)
		if err != tm.err {
			t.Log(tm.input)
			t.Errorf("index: %d got: %s wanted: %s", index, err, tm.err)
		}
		if err != nil {
			continue
		}
		if v0.String() != tm.output {
			t.Errorf("index: %d got: %s wanted: %s", index, v0.String(), tm.output)
		}

		if v0.entityDomain != tm.entityDomain {
			t.Errorf("index: %d got: %q wanted: %q", index, v0.entityDomain, tm.entityDomain)
		}
		if v0.entityType != tm.entityType {
			t.Errorf("index: %d got: %q wanted: %q", index, v0.entityType, tm.entityType)
		}
		if v0.decoded != tm.decoded {
			t.Errorf("index: %d got: %q wanted: %q", index, v0.decoded, tm.decoded)
		}
	}
}

type generateTestCase struct {
	realm        string
	entityDomain string
	entityType   string
	extrinsicID  string
	expected     string
	err          error
}

func TestGenerateV0(t *testing.T) {
	var testmap = []generateTestCase{
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			expected:     "blox0.infra.host.us-com-1.",
			err:          ErrEmptyExtrinsicID,
		},
		{
			realm:        "us-com-2",
			entityDomain: "infra",
			entityType:   "host",
			expected:     "blox0.infra.host.us-com-2.",
			err:          ErrEmptyExtrinsicID,
		},

		// ensure `=` is not part of id when encoded
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "1",
			expected:     "blox0.infra.host.us-com-1.ivmfiurreaqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "12",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "123",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgizsaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "1234",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztiiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "12345",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinja",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "123456",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjweaqcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "1234567",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg4qcaiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "12345678",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg44caiba",
		},
		{
			realm:        "us-com-1",
			entityDomain: "infra",
			entityType:   "host",
			extrinsicID:  "123456789",
			expected:     "blox0.infra.host.us-com-1.ivmfiurrgiztinjwg44dsiba",
		},
	}

	for index, tm := range testmap {
		v0, err := NewV0("",
			WithEntityDomain(tm.entityDomain),
			WithEntityType(tm.entityType),
			WithRealm(tm.realm),
			WithExtrinsicID(tm.extrinsicID),
		)
		if err != tm.err {
			t.Logf("test: %#v", tm)
			t.Errorf("index: %d got: %s wanted error: %s", index, err, tm.err)
		}
		if err != nil {
			continue
		}

		if v0 == nil {
			t.Errorf("unexpected nil id")
			continue
		}

		//		t.Log(v0)
		//		t.Logf("%#v\n", v0)

		validateGenerateV0(t, tm, v0, err)

		parsed, err := NewV0(v0.String())
		if err != tm.err {
			t.Logf("test: %#v", tm)
			t.Errorf("got: %s wanted: %s", err, tm.err)
		}
		if err != nil {
			continue
		}

		validateGenerateV0(t, tm, parsed, err)
	}
}

func validateGenerateV0(t *testing.T, tm generateTestCase, v0 *V0, err error) {
	if len(tm.extrinsicID) > 0 {
		if v0.decoded != tm.extrinsicID {
			t.Errorf("got: %q wanted decoded: %q", v0.decoded, tm.extrinsicID)
		}
		if v0.String() != tm.expected {
			t.Errorf("got: %q wanted bloxid: %q", v0, tm.expected)
		}
	} else {
		if strings.HasPrefix(tm.expected, v0.String()) {
			t.Errorf("got: %q wanted prefix: %q", v0, tm.expected)
		}
		if len(v0.decoded) < 1 {
			t.Errorf("got: %q wanted non-empty string in decoded", v0.decoded)
		}
	}

	if -1 != strings.Index(v0.String(), "=") {
		t.Errorf("got: %q wanted bloxid without equal char", v0.String())
	}

	if v0.Realm() != tm.realm {
		t.Errorf("got: %q wanted realm: %q", v0.Realm(), tm.realm)
	}
	if v0.Domain() != tm.entityDomain {
		t.Errorf("got: %q wanted entity domain: %q", v0.Domain(), tm.entityDomain)
	}
	if v0.Type() != tm.entityType {
		t.Errorf("got: %q wanted entity type: %q", v0.Type(), tm.entityType)
	}
	if len(tm.extrinsicID) > 0 {
		if v0.decoded != tm.extrinsicID {
			t.Errorf("got: %q wanted extrinsic id: %q", v0.decoded, tm.extrinsicID)
		}
	}
}
