// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

// LoadServerTLSConfig returns a server TLS config by loading the certificate file and the key file.
func LoadServerTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	keyPEMBlock, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return ServerTLSConfig(certPEMBlock, keyPEMBlock), nil
}

// LoadClientTLSConfig returns a client TLS config by loading the certificate file.
func LoadClientTLSConfig(certFile, serverName string) (*tls.Config, error) {
	certPEM, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	return ClientTLSConfig(certPEM, serverName), nil
}

// ServerTLSConfig returns a server TLS config by the certificate data and the key data.
func ServerTLSConfig(certPEM []byte, keyPEM []byte) *tls.Config {
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

// ClientTLSConfig returns a client TLS config by the certificate data.
func ClientTLSConfig(certPEM []byte, serverName string) *tls.Config {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		panic("failed to append certificates")
	}
	return &tls.Config{RootCAs: certPool, ServerName: serverName}
}

// DefalutTLSConfig returns a default server TLS config.
func DefalutTLSConfig() *tls.Config {
	return ServerTLSConfig(DefaultCertPEM, DefaultKeyPEM)
}

// SkipVerifyTLSConfig returns a client TLS config which skips security verification.
func SkipVerifyTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

// DefaultKeyPEM represents the default private key data.
var DefaultKeyPEM = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDa3lfVcQZg3Ra2
DCeTPM9I8cv35Y+R4niXJ7c2U9TvGE3l8zfsLBXtdN4bSlmaimOnOmfx0aVJ8XwL
qcIMspJmzG9UlGdlOfirMTYCybvwhEf9bZc9lmLv27C4++4IljNF9sSv/Lnbdl5V
Nr+lY1xKRR3HPpwwuJj3jh3TznzAnb0QnIRTyGGVThyE6uUQAgx8/taGenJDkzb7
pry4kRvz+GgjAvhi/KOgxho7G6PLfzXeS+iPyaMg5npd3B90XIzaaXr4/yffC5BU
bynHhZLmWKJXSp7brjiZzpFV8np7wKYrtqXW4My2MtMASnvfXrCfTwQ3FU0biBsk
7dQHCcuzAgMBAAECggEAdN6zQhsHT+Pew7j7zOh0uzu6MZYYMssen4Aqmczr7/wn
ZHmaS/dCgjicfTAXZqktC1fptzu+KhzToxqzrroP6OqTLDPOfkQVX7x4XcbBH25T
TqUdVFqgW/oQhMap1VX27Q4W+u5VhDXRq2j/rt2+oz4C56isGGwJ6m6tyLMC9IqJ
Ul9fHrLKKjHltYkCMYzbUP/9QVs9yMlw04BbxCvML21s3ikNuGc8qdQhoHkmxXns
zUR9+P7CkMhSvhojs7MVgaflGozNna89MYAgX+0mCGkWqOXEoFN3n4HdxwW1nBHC
34YndQdOsViO7j9o1SJMOBLMXiQexH+YDJMvjZpsEQKBgQDynyxfINLnHUz1Wo8K
Z1dwmP+dd2av/MbBVsEEyxAugLW8a7Ks6bDlxk8VKB4GAkqzx6Ap+YywZDRJewKn
XUoEG8TPo4dZBy3ttyXTk240zDi/NIJtVRhGxeOGX8zcmwtGHjq694RYeCDrMDWp
yRCJHVUSYUhHwtVwvSZK8JSKCwKBgQDm798FKiqh8UFyuOwKkfdzPYjM6fDg3JuR
E7kmyaeFRz1X0c9zZcWE+ehf9nZnwfU04ZL1WIrjkYWUBxcrBFjChAmFKqFOf97m
0w8jCifuBu+AzaSfW39rzcbpCof9GIHTEczGIjIbj62NQfxhKejP/eA///o81Cf2
hSnUpjn1+QKBgE0jyLLSN9wdl8tmqJYRN17odlU1kmOgBf2QvLvuaE2wxJeM0nlh
r8nOnHRIlgspDWFNtiHCYzXuFiXKw5Q89/yIa7Hs92qZ+sNa+N7lQCPvTpeUdWeX
p6lQ379olDUL4rC/icLKUbzjLOw6HsXF1MkTl2nJnnaafsxih1tKVJ/zAoGAEed8
+fCH96A1u8g8fKFOdv/JUGG+zCAua3QFAc3WkA2y4tEgbUjxpFqfunjoOykdcqke
dKkVs4j/uzdFg49Ftmb4OfvRH73oMSsh3EyYResBvJG09qnoWhpNFpo7atLwlcWm
g6H5Eov0H6SDBaFzLFT5gty8sOSd6I3wbU0p5zkCgYAwQe8+M7Su2v3mA0vbxbGb
W96El5n15YRa6JHOigC+5mBhXilnDE8qomFkfELDOnQ+hdkgqbFd7P1/+K5raV+I
aGh+dZd2MKnLevVoMexu40NQLVyJTOqumG05NNgmfg7VE8QUbXKfz+9pmfYFSZGS
Wx4EqMDhdG9wlTsHGb1I/Q==
-----END PRIVATE KEY-----
`)

// DefaultCertPEM represents the default certificate data.
var DefaultCertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDfDCCAmQCCQCAHkBfX03BnTANBgkqhkiG9w0BAQsFADB/MQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxEDAOBgNVBAcMB0JlaWppbmcxDjAMBgNVBAoMBUhTTEFN
MQwwCgYDVQQLDANSJkQxEjAQBgNVBAMMCWhzbGFtLmNvbTEfMB0GCSqGSIb3DQEJ
ARYQNzkxODc0MTU4QHFxLmNvbTAgFw0yMDA5MjMwMzE3NTdaGA8yMTIwMDgzMDAz
MTc1N1owfzELMAkGA1UEBhMCQ04xCzAJBgNVBAgMAkJKMRAwDgYDVQQHDAdCZWlq
aW5nMQ4wDAYDVQQKDAVIU0xBTTEMMAoGA1UECwwDUiZEMRIwEAYDVQQDDAloc2xh
bS5jb20xHzAdBgkqhkiG9w0BCQEWEDc5MTg3NDE1OEBxcS5jb20wggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDa3lfVcQZg3Ra2DCeTPM9I8cv35Y+R4niX
J7c2U9TvGE3l8zfsLBXtdN4bSlmaimOnOmfx0aVJ8XwLqcIMspJmzG9UlGdlOfir
MTYCybvwhEf9bZc9lmLv27C4++4IljNF9sSv/Lnbdl5VNr+lY1xKRR3HPpwwuJj3
jh3TznzAnb0QnIRTyGGVThyE6uUQAgx8/taGenJDkzb7pry4kRvz+GgjAvhi/KOg
xho7G6PLfzXeS+iPyaMg5npd3B90XIzaaXr4/yffC5BUbynHhZLmWKJXSp7brjiZ
zpFV8np7wKYrtqXW4My2MtMASnvfXrCfTwQ3FU0biBsk7dQHCcuzAgMBAAEwDQYJ
KoZIhvcNAQELBQADggEBAA4rrtWczvjVpttxJ7pbXQlmvVrakPwqqKEQ09hxcoqY
EKkCucjJwFFQi1fNQBKpb+3BwlHIcfqdwpURiTwQjPmRgVhqdFqHE5pNF9EXdNm7
zaylUiu+ySKKHHnCVagM7UszovCoRYY3hq75UsGwR+9WWxOoWRz43NdOTBBDE9y7
JkRowySk9JE5isec+G0tDf6Fyj/3zWshWQalEH/Aq1Af0BMtWQL4VYXbealqK6rq
MOwPd7m67gCJlNREX2JnMDBM2A9QcAIzhYrHBx5w6UhUwSL6IFhJzdFXl4klsKUQ
cmw7rbPxsuPIyPlCobdtFoVpFN5vnOnF42nCb8tr0Xs=
-----END CERTIFICATE-----
`)
