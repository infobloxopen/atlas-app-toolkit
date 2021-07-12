package bloxid

import (
	"testing"
)

func TestNewV0(t *testing.T) {
	var testmap = []struct {
		input      string
		output     string
		entityType string
		decoded    string
		err        error
	}{
		{
			"",
			"",
			"",
			"",
			ErrIDEmpty,
		},
		{
			"blox0.infra_host..zdud52youveke5sovyoc66cjxw3l55jc",
			"blox0.infra_host..zdud52youveke5sovyoc66cjxw3l55jc",
			"infra_host",
			"c8e83eeb0ea548a2764eae1c2f7849bdb6bef522",
			nil,
		},
		{
			input: "bloxv0.infra_host..asdfasdfasdfasdfasdfasdf",
			err:   ErrInvalidVersion,
		},
		{
			input: "blox0...adsf",
			err:   ErrInvalidEntityType,
		},
		{
			input: "blox0.a..a",
			err:   ErrInvalidUniqueIDLen,
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

		if v0.decoded != tm.decoded {
			t.Errorf("got: %q wanted: %q", v0.decoded, tm.decoded)
		}

		if v0.entityType != tm.entityType {
			t.Errorf("got: %q wanted: %q", v0.entityType, tm.entityType)
		}
	}
}

func TestGenerateV0(t *testing.T) {
	var testmap = []struct {
		shortid    string
		entityType string
		output     string
		err        error
	}{
		{
			shortid:    "",
			entityType: "infra_host",
		},
	}

	for _, tm := range testmap {
		v0, err := GenerateV0(&V0Options{
			EntityType: tm.entityType,
		}, WithShortID(tm.shortid))
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
		t.Log(v0)
		t.Logf("%#v\n", v0)
	}
}
