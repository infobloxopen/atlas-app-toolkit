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
		{
			"",
			"",
			"",
			"",
			"",
			ErrIDEmpty,
		},
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

	for _, tm := range testmap {
		v0, err := NewV0(tm.input)
		if err != tm.err {
			t.Log(tm.input)
			t.Errorf("got: %s wanted: %s", err, tm.err)
		}
		if err != nil {
			continue
		}
		if v0.String() != tm.output {
			t.Errorf("got: %s wanted: %s", v0.String(), tm.output)
		}

		if v0.entityDomain != tm.entityDomain {
			t.Errorf("got: %q wanted: %q", v0.entityDomain, tm.entityDomain)
		}
		if v0.entityType != tm.entityType {
			t.Errorf("got: %q wanted: %q", v0.entityType, tm.entityType)
		}
		if v0.decoded != tm.decoded {
			t.Errorf("got: %q wanted: %q", v0.decoded, tm.decoded)
		}
	}
}

func TestGenerateV0(t *testing.T) {
	var testmap = []struct {
		realm          string
		entityDomain   string
		entityType     string
		output         string
		expectedPrefix string
		err            error
	}{
		{
			realm:          "us-com-1",
			entityDomain:   "infra",
			entityType:     "host",
			expectedPrefix: "blox0.infra.host.us-com-1.",
		},
	}

	for _, tm := range testmap {
		v0, err := GenerateV0(&V0Options{
			EntityDomain: tm.entityDomain,
			EntityType:   tm.entityType,
			Realm:        tm.realm,
		})
		if err != tm.err {
			t.Errorf("got: %s wanted: %s", err, tm.err)
		}
		if err != nil {
			continue
		}

		if v0 == nil {
			t.Errorf("unexpected nil version")
			continue
		}

		if strings.HasPrefix(tm.expectedPrefix, v0.String()) {
			t.Errorf("got: %q wanted prefix: %q", v0, tm.expectedPrefix)
		}
		t.Log(v0)
		t.Logf("%#v\n", v0)
	}
}
