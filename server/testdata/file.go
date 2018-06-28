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

// WritePEMFiles creates a tmp dir with ca.pem, server.pem, and server-key.pem, client.pem, client-key.pem and the func to remove it
func WritePEMFiles(dir string) (string, func(), error) {
	tempDir, err := ioutil.TempDir(dir, "go-test-pemfiles")
	if err != nil {
		return "", nil, err
	}

	data := `-----BEGIN CERTIFICATE-----
MIIC5zCCAc+gAwIBAgIBATANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwptaW5p
a3ViZUNBMB4XDTE3MDgxNDE5NTcyM1oXDTI3MDgxMjE5NTcyM1owFTETMBEGA1UE
AxMKbWluaWt1YmVDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQx
IIST+MO/bT4bqgPVWecRMHmoasELlqVWXxDIO+zKgzPqtM+NvUrMyQnV7kVaCbWn
U7QIf5rV2dc3TbjCYPb7FK9bAlyHXDX0gcYkLZ3CPARXsS9g4XaeeyO73nQjriho
HoNS/aMQ7Xo0BdW258Il1ssx67GR/JYCKtny32yCWmYc5XRayvO1yMH0zSSW5CeF
4OngZF/rcrKhZSNG+j2FMCKm4yQU/a83ogz5jWaBzbQ6pXoBpADNjeHg/YF57Q/X
Ovu8gaHNHfFvVSPETTouGlS9CYorQzCjwJJ88+/wonOTaxryhzsGSw1vCFircFli
PyCC8wzLWeU7pR3xmeECAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMB0GA1UdJQQW
MBQGCCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4IBAQA2kEWfJvifeakdbnf8sLWhn34v9ESToPE06vZs4o51sk9wzQ+h
EE6xWj78/rUPJPwd4ke+3IWWMQrqShnTE7DgVyjwrjNBUTN+TS0hqrqlMXTK1Ctk
iqMEvklOHDrASDWu9FZ3spn4A2VOCeQcOll6oqfljeDTYcXv5VVOSIWMPxxjF+Pp
50H3cooLyXb393C+fJmyahy1VWL+S+Cj/vxCBPbag8yIS0qhXfOfzNcKB40og9QI
1Ap9hEZjmfCuPgBGH+rw124bT5EkwH3hTEn3jjmzdpzQtZjqQouefs+ZHBYGbCg0
tDFyB8yeacrDVieDxSAIjFsYcY3F2zzj3BRO
-----END CERTIFICATE-----`
	path := filepath.Join(tempDir, "ca.pem")
	if err := ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}
	data = `-----BEGIN CERTIFICATE-----
MIIDLTCCAhWgAwIBAgIUeVPFYUf32wT8z06H8zMW6fMzfwwwDQYJKoZIhvcNAQEL
BQAwFTETMBEGA1UEAxMKbWluaWt1YmVDQTAeFw0xODA2MjgwMTIxMDBaFw0xOTA2
MjgwMTIxMDBaMBoxGDAWBgNVBAMTD3Rsc3Rlc3Qtc2VydmljZTCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBAM+XRKlxDGPw8UP8/7Hj5XBjepIhLiMcTCOK
XSZv6Oh+akfYMFdEwNoLJC3Pff3U+vs128NyQbqc+8dpYogtH5CkwCQeA1Bi+ndz
AeaGiwUhWQ5M+BiNDkL2r1kA2Nm4ngRn44nIj1E7wazhPQY49OTD0ylfR6Ke5j/B
Y+K6vSfzw9uD1Jqob6XXG+8THIYhU8vnJuiPrp69gMdm585HMzb4aOa2DCLImYyP
ZaahMb1ad2N4tlC74P7nxE3nAP7+RBdO+EpcjkNdjYRYo6ZHZTxW4BzwIvKrumFn
wei0TiZDMbYdydnEG9QBv3DhVxI2Vi6Qfmiy/urHk9jjx7Kz79cCAwEAAaNwMG4w
DgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQC
MAAwHQYDVR0OBBYEFPFKoiWH7vByo1ycFb/vDrIhF5l8MBoGA1UdEQQTMBGCCWxv
Y2FsaG9zdIcEfwAAATANBgkqhkiG9w0BAQsFAAOCAQEAiHgVcA8Yi1zETM/1bIqT
5ktpOh+xvJIfgXOoeW4ymjKH00bNeTuEbAPm7lQSx2HZhvjpiwguSSLkkZ7AJihX
IyomgyleWq1BQmlqSZE/vcQO3CzOYxpMPx3QnbZjhcOAQRnmhPCDUSyt0bjYu9yO
t/nYvJeb+xNKUNVR6uj9brLDrzfFqT3Tnyz2ycovkqTFtylenwahSietbmw6JxHT
wLocWi2GxowTb7as/XqNWo0hV7isHOLyhHvtxVsWlMqvy7et2QRkOgqY7QI+GABp
c3HChwTs75g6ZyBGghscJGxxQBcMuAm0DEM+NpjfejUV/dTe/a3NymuUt4aiuPG+
0w==
-----END CERTIFICATE-----`
	path = filepath.Join(tempDir, "server.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	data = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAz5dEqXEMY/DxQ/z/sePlcGN6kiEuIxxMI4pdJm/o6H5qR9gw
V0TA2gskLc99/dT6+zXbw3JBupz7x2liiC0fkKTAJB4DUGL6d3MB5oaLBSFZDkz4
GI0OQvavWQDY2bieBGfjiciPUTvBrOE9Bjj05MPTKV9Hop7mP8Fj4rq9J/PD24PU
mqhvpdcb7xMchiFTy+cm6I+unr2Ax2bnzkczNvho5rYMIsiZjI9lpqExvVp3Y3i2
ULvg/ufETecA/v5EF074SlyOQ12NhFijpkdlPFbgHPAi8qu6YWfB6LROJkMxth3J
2cQb1AG/cOFXEjZWLpB+aLL+6seT2OPHsrPv1wIDAQABAoIBAHLaC0f/3s4QiTuH
Z2RhZRiYQUFGEEOmrU1giQbyFZdEEcMsDhrKVDSOw3aA/QEJ684+zxbESS9ZGUgL
u1MCPBuHuzKAVG8OQ+sAe0WynIm7GI178iuLJx/PYdZJTLCwnuRrIg2iJADaod3n
RB8ENiG3YkXajEShy1vswDm5/VtzWIaKIvSYUNBvOQhPdpR/GSigToIaNqS4Z/RW
VWvQullTh97ZqvvTM9B544PFzRnEHHpxHB5QipwLp2SqUAIVZhP4gjm6cVSCefYe
Pd9jhL42nhQHwgRwRf5yXiYVX2AErrXwq9cVw/myBUAfJZ+BJ9kiO47P5Mkq8tUr
uf6qV9kCgYEA6XSlUH88WqxG7tNLhWQ+NFBXKaOTqsJM+LwMW3WeNyZSgVBoRnS/
QQpW46ji7kY2GFlLdc0Xim/SRJvCJ4biTkl/QnYRrufELJDd5WoN9n8+Tuim3B9v
tsl+cE3RHp0e/A0ULe5IdzX4uHfi1Ss4EJWJIHhwMkVvpl4KhP7f6DUCgYEA46M1
hE6ECrCSyQSImVewZ3NOP/G3j1+XdtyBqDK+wep3e06zim7oriGjQLTy+gxYoxVK
6I+VVeJQ8Y4qqZVUOQIscGdMMw0W/Efpj9gytdEWlpgAD5One61fSQ8RfbY6bFv2
5MqFagGimoZx3iQH/50mxC+KE30YL+cJeWrJcVsCgYEAkEMGjPGzKAzhYF+tgWZq
kgU7d32fmJus2N/LexD5jfbecQ5xAWjPbq+m9dO9N6SndPBpEwiDjYaAFulxVt+h
JOOCAl3Xm4+YyDlVBZk9u57xr+1QfyHl9Lwap+dOXG6XYQXr/F4M5a2yXrumrjeg
0460SB5kpowF7HacZYbicikCgYAGq0FmHua/aWzjdr6Jv4frf/VK6kn2aVaGpO8n
flUYWUYm1qdr5tPqRhICU0rLCJGQNY98QLifS4ITkZauYTGWefnTUTNqS3fg7Dpr
fGn/6aA/yTQ3QJwng2zHNynMBQqxIgCZs1U1Rdb9r/KmD2gslO4N0Va6O2/590rP
w9EjAQKBgQDl5Rn2cEiGUWMOJr+zWEMCneIJDbba7H6lg6Tc9ARx1DVrlkhq1Fz7
vjfq2cvmCeFMHWaaa24+i6XC820Zm9SJpmNioZDVFD6j4BKmumPhH2zR3PoRlYRc
GCcT/8NMy2cXWboQOHtiUjH1D6ccQF5Jc6XQKUF6I6Fqx9uRhSHofg==
-----END RSA PRIVATE KEY-----`
	path = filepath.Join(tempDir, "server-key.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	data = `-----BEGIN CERTIFICATE-----
MIIDEDCCAfigAwIBAgIUXGiqBNoD/mSxl3avDaUCm2JiOEowDQYJKoZIhvcNAQEL
BQAwFTETMBEGA1UEAxMKbWluaWt1YmVDQTAeFw0xODA2MjgwMTIxMDBaFw0xOTA2
MjgwMTIxMDBaMBkxFzAVBgNVBAMTDnRsc3Rlc3QtY2xpZW50MIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3tQVilGYr49LiW8d5cWb6pAlB0ioudSHdxdW
schsZwiPxhjWEKxZyxspuXoGMQQ9L6AK+Fo+jdFbeJDfsQeTX2TDtVjKLNRPBFgr
m/9Qy+ODUHNP63sOvcfqSAX3JKHWuvphZq1pVT0A9du6FegWFdz2P+V6Sr1rKBuX
kunw+2wicdYzUwfp2IUc+7O/MfnKs7c1BgeF9spFx9UDapvk8l/SNtAAmeD2h51w
BwfgT640OXMumi0sHien8JW/mRM/6sP2FmdUFZb3q/e9qO37JJbydMhqCyZeqAcU
ZRyOrmQ5NZHaCvxTNUX2r1xOTRPO7cPWdOY1P8snjHucs9yY1QIDAQABo1QwUjAO
BgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwIwDAYDVR0TAQH/BAIw
ADAdBgNVHQ4EFgQU7g3WJ7t96LRV/+pxpuF8sY4HL50wDQYJKoZIhvcNAQELBQAD
ggEBAHn58TsdmdzN9ZWpaqrvXs8yABJj2FSApf1bRHEGMzl2Iq7ouz3nb47w11y1
dqOw+I2Am9Ru5pfT10vVsoyUYcccdWGk5xZTyakNknrIqtKGIJal+C7U2AV7eEzZ
yNuoJXcmuSzwYXmUHwn/Ut2bYd2GU4Wl96CkV+5sIAX8JJeyn35M3ByLkSRkD9+9
DZLgw4ijKQYlpRLv5RssNorx423NSpE8G2hTTSzsQ6ln5WUeT5G9nj9nzeGtl1u2
523/r87f5RSIMEYQlUMOD5pyF1S9t9JH0YsWVJEpZgrPmoQfK/40K5ccwYlCbOnc
iPKLg3KeqrmbrNUaoCRmS/OkV2c=
-----END CERTIFICATE-----`
	path = filepath.Join(tempDir, "client.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	data = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA3tQVilGYr49LiW8d5cWb6pAlB0ioudSHdxdWschsZwiPxhjW
EKxZyxspuXoGMQQ9L6AK+Fo+jdFbeJDfsQeTX2TDtVjKLNRPBFgrm/9Qy+ODUHNP
63sOvcfqSAX3JKHWuvphZq1pVT0A9du6FegWFdz2P+V6Sr1rKBuXkunw+2wicdYz
Uwfp2IUc+7O/MfnKs7c1BgeF9spFx9UDapvk8l/SNtAAmeD2h51wBwfgT640OXMu
mi0sHien8JW/mRM/6sP2FmdUFZb3q/e9qO37JJbydMhqCyZeqAcUZRyOrmQ5NZHa
CvxTNUX2r1xOTRPO7cPWdOY1P8snjHucs9yY1QIDAQABAoIBADGoV+iMSJePOok1
Lxd+k0GRv/7AjYFkQJAkrlhOTwIjIU9HR6VNk3E063Z/IBQaWrxbUTaJfflC17yl
pIQiKRCQNyEZ2WxmH/na4FUSb+IQILp0CDJ1KRLYE3jbg3kxa9YdaElmidvKDYW8
4tpw70MOC/6vWDKBxfiZNY0y/1Y4F3v7zlHniFJx6ogUNe1kafefQJkG0dVpAlv4
FvSCUafPuR18kw8glQiL97XsuuMlvameXMJGdv45D342YtFilN7M3j1xeSHy1+WP
XDTpkrjg0q3J33Iiiu8Bo3o/j7YX8KY7ixumXKaKjTnMedhcgCly5eR2p2gjZYZF
6Z6CiSECgYEA7LCdgaFBtokszLP61FRDatjDlUvtY9AnLARk7wVbgJd73VMLyVBS
SzRjjO9edVaEB83K/nGBLwTEK9Dswqv12+Q4a2T4/9nLa+irHGMw6pgwHcbmSf39
jps5ODkiirgXc+AuMGMk0qh//33Qsd3Eqls4/sBk3MVofG5E6enIatkCgYEA8QH3
OIbZoqHYo2Yw4XFdRaZhbz0rxO6Dsa5cahAlqJcX5OH76UG++WlCVMhSPFUxzTfl
SPM96BaZxlnbIsGVzicDEO+7O53Uu8QDdCHlqp0hsIr1Fi7QWLDJ0e/xE6frcOWg
N4IhXGQNONSmoNrSQzH17ybMH5XEmsmY4EAPCF0CgYEAl6UstPIhTRckSd8CVOnL
6/gHj27/IJUrk8sY8/8lugTUSmA7y/aXUzG0moZ+qYUNwIY8ibslPn+6RCxulOdh
9UmKUFx4IExlRbTjdKOkoplxMpLN1xhRTP3ssjYBCImcFRTL4xqSbBmjMIlmnZ7t
swwRPz77IGumXxqzMn8jdjkCgYBqBiJbJL/Tkv26DHfOhc+xl1tf03pQ3VjkLr+L
DWVzwFyLnXr0B69bC5pZr/K1hgktrbxZlmCSnHaz0s3bgWxEz9bCeaRVur5eiAG4
8jyWDSBICSl+w8N2cPeoOrVEn2etN+d+4+mHOqCycqKHOxyq4Oy/c8Ly1jEyoyN0
69lxJQKBgQDnmef6rVUfq4WZtoKGo3zjEbNkTHz3TjmL/AyikCG/nboWrcNvibhl
oRi+k+/a1qZP85kv1DxxEVnso/aqMMv/MFeyqEnV2ljLjARwi9AWzapyDPz5Vbsw
XeNxjzQ6ZUGL7bmCF3BFC0biRgv+bE/XMkswbmw47jHaYI9ddaMFWw==
-----END RSA PRIVATE KEY-----`
	path = filepath.Join(tempDir, "client-key.pem")
	if err = ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		return "", nil, err
	}

	rmFunc := func() { os.RemoveAll(tempDir) }
	return tempDir, rmFunc, nil
}
