package servertest

import (
	"net"
)

// NewLocalListener returns a system-chosen port on the local loopback interface, for use in end-to-end HTTP tests
func NewLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}
