package resource

import "testing"

func TestNil(t *testing.T) {
	tcases := []struct {
		Identifier *Identifier
		Expected   bool
	}{
		{
			Identifier: nil,
			Expected:   true,
		},
		{
			Identifier: &Identifier{},
			Expected:   true,
		},
		{
			Identifier: &Identifier{ResourceId: "uuid"},
			Expected:   false,
		},
	}

	for n, tc := range tcases {
		if v := Nil(tc.Identifier); v != tc.Expected {
			t.Errorf("%d: invalid result %t, expected %t", n, v, tc.Expected)
		}
	}
}
