// Adapted from https://github.com/coredns/coredns
package server_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// TempFile will create a temporary file on disk and returns the name and a cleanup function to remove it later.
func TempFile(dir, content string) (string, func(), error) {
	f, err := ioutil.TempFile(dir, "go-test-tmpfile")
	if err != nil {
		return "", nil, err
	}
	if err := ioutil.WriteFile(f.Name(), []byte(content), 0644); err != nil {
		return "", nil, err
	}
	rmFunc := func() { os.Remove(f.Name()) }
	return f.Name(), rmFunc, nil
}

// WritePEMFiles creates a tmp dir with ca.pem, cert.pem, and key.pem and the func to remove it
func WritePEMFiles(dir string) (string, func(), error) {
	tempDir, err := ioutil.TempDir(dir, "go-test-pemfiles")
	if err != nil {
		return "", nil, err
	}

	data := `-----BEGIN CERTIFICATE-----
MIICEjCCAXugAwIBAgIRAK9oivV13n8NjkrxlRObpfQwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEAq03JACUKtgXTKoYNvFPEKmIk5fS4x2MxczPfiT8KLo2gVikfEMqCtoIt
NcXL+xxYZ8dA2Y26Yk+WjeEzB+/W1qYbei6kZR+GOy3TFINJoqYFZq4sDF6c1Gch
ACqB4oE+4kLdq4hS9cM2IjEUovBQa+Q9frU7ONLLFfOWwJ5Wt0ECAwEAAaNmMGQw
DgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQF
MAMBAf8wLAYDVR0RBCUwI4IJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAA
AAABMA0GCSqGSIb3DQEBCwUAA4GBAHzFYeTxkJdvcahc7C1eKNLkEnus+SBaMeuT
QSeywW57xhhQ21CgFAZV2yieuBVoZbsZs4+9Nr7Lgx+QuE6xR3ZXOBeZVqx3bVqj
jc5T1srmqkU/gF/3CALuSuwHFyCIdmuYkgmnDUqE8vJ4eStuDaMVWjGvPYmi3am7
yc1YAUB7
-----END CERTIFICATE-----`
	path := filepath.Join(tempDir, "ca.pem")
	if err := ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}
	data = `-----BEGIN CERTIFICATE-----
MIICEjCCAXugAwIBAgIRAK9oivV13n8NjkrxlRObpfQwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEAq03JACUKtgXTKoYNvFPEKmIk5fS4x2MxczPfiT8KLo2gVikfEMqCtoIt
NcXL+xxYZ8dA2Y26Yk+WjeEzB+/W1qYbei6kZR+GOy3TFINJoqYFZq4sDF6c1Gch
ACqB4oE+4kLdq4hS9cM2IjEUovBQa+Q9frU7ONLLFfOWwJ5Wt0ECAwEAAaNmMGQw
DgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQF
MAMBAf8wLAYDVR0RBCUwI4IJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAA
AAABMA0GCSqGSIb3DQEBCwUAA4GBAHzFYeTxkJdvcahc7C1eKNLkEnus+SBaMeuT
QSeywW57xhhQ21CgFAZV2yieuBVoZbsZs4+9Nr7Lgx+QuE6xR3ZXOBeZVqx3bVqj
jc5T1srmqkU/gF/3CALuSuwHFyCIdmuYkgmnDUqE8vJ4eStuDaMVWjGvPYmi3am7
yc1YAUB7
-----END CERTIFICATE-----`
	path = filepath.Join(tempDir, "cert.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	data = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCrTckAJQq2BdMqhg28U8QqYiTl9LjHYzFzM9+JPwoujaBWKR8Q
yoK2gi01xcv7HFhnx0DZjbpiT5aN4TMH79bWpht6LqRlH4Y7LdMUg0mipgVmriwM
XpzUZyEAKoHigT7iQt2riFL1wzYiMRSi8FBr5D1+tTs40ssV85bAnla3QQIDAQAB
AoGABtvWcGsLQr549froEeJIuGm1kH975n/SOwqYqKYdgj+pa8m5tLJnCWes57pD
sIox//W6YvuJuuX04TljEa5Iq7604Ien0x/FCCQshW/3/skEXkKc89+a1eLw9wt/
c75qow5S5CG01Ht/+AqWCzkSADE/QTFfnSMLfYGfOm1X7AECQQDWAtGny7GGeBH+
C/nMLags2q0nc0ZZ/QcdwMGtN2q0ZfiYhQw968FuEiWSeiiGhGUPTrOkERU/l93S
NYrovJkNAkEAzOnnTdYWwmfs+LBQIYGQOmuTYbmzn0lpmeDUsCtSi3G+pRVCvpoc
4sFMwrFTea1257fryUfxXUkE5mGYYqQiBQJAW6VvZNzc1AndIp68RUyUBUlL92Xt
DaJGht5B0ky1/DTixWXMfUPVXK6WumhnrFtL78czNKJAKDB/xII7TzlcjQJBALhD
2fj3fM3i0IitW9FVhhHSrNyjNjAVvv1d3URyIK8+YJZosPVe9ny+ID2vYgY4A4XJ
sSD2LciaIerddj+1otUCQQDNLXTkZ2riEEhNoZfiDUumlJgAJw0M07SFyKyU60yn
r3nPX1rJpUYnyRYsRf+F6dwvAqECKgQao/QRKriAubDk
-----END RSA PRIVATE KEY-----`
	path = filepath.Join(tempDir, "key.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	rmFunc := func() { os.RemoveAll(tempDir) }
	return tempDir, rmFunc, nil
}
