package integration

import (
	"fmt"
	"net"
)

const (
	portRangeMin = 0
	portRangeMax = 65535
)

var (
	errPortMax = portError{
		message: fmt.Sprintf("port cannot be greater than %d", portRangeMax),
	}
	errPortMin = portError{
		message: fmt.Sprintf("port cannot be less than %d", portRangeMin),
	}
	errPortNotFound = portError{
		message: fmt.Sprintf("no open port found %d", portRangeMin),
	}
)

type portError struct{ message string }

// Error returns a string with the port-related error message
func (p portError) Error() string {
	return fmt.Sprintf("port discovery error: %s", p.message)
}

// GetOpenPortInRange finds the first unused port within specific range
func GetOpenPortInRange(lowerBound, upperBound int) (int, error) {
	if lowerBound < portRangeMin {
		return -1, errPortMin
	}
	for lowerBound <= portRangeMax && lowerBound <= upperBound {
		if _, err := net.Dial("tcp", fmt.Sprintf(":%d", lowerBound)); err != nil {
			return lowerBound, nil
		}
		lowerBound++
	}
	if upperBound > portRangeMax {
		return -1, errPortMax
	}
	return -1, errPortNotFound
}

// GetOpenPort searches for an open port on the host
func GetOpenPort() (int, error) {
	return GetOpenPortInRange(portRangeMin, portRangeMax)
}
