package resource

import "testing"

func TestValid(t *testing.T) {
	tcases := []struct {
		Identifier *Identifier
		Expected   bool
	}{
		{
			Identifier: nil,
			Expected:   false,
		},
		{
			Identifier: &Identifier{ResourceId: "uuid"},
			Expected:   true,
		},
	}

	for n, tc := range tcases {
		if v := Valid(tc.Identifier); v != tc.Expected {
			t.Errorf("%d: invalid validation result %t, expected %t", n, v, tc.Expected)
		}
	}
}
