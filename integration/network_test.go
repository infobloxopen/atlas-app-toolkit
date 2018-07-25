package integration

import (
	"testing"
)

func TestGetOpenPortInRange(t *testing.T) {
	var tests = []struct {
		name     string
		minRange int
		maxRange int
		err      error
	}{
		{
			name:     "negative port range",
			minRange: -1,
			maxRange: 0,
			err:      errPortMin,
		},
		{
			name:     "exceed maximum port upperbound",
			minRange: portRangeMax + 1,
			maxRange: portRangeMax + 10,
			err:      errPortMax,
		},
		{
			name:     "exceed specified port uppercbound",
			minRange: 20,
			maxRange: 10,
			err:      errPortNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := GetOpenPortInRange(test.minRange, test.maxRange); err != test.err {
				t.Errorf("unexpected error when getting open port: have %v; expected %v",
					err, test.err,
				)
			}
		})
	}
}
