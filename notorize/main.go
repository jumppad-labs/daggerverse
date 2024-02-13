package main

import (
	"context"
)

type Notorize struct {
	P12Cert         *File
	P12CertPassword *Secret
	NotoryKey       *File
	NotoryKeyID     string
	NotoryIssuer    string
}

// WithP12Cert sets the p12 certificate file and password
func (m *Notorize) WithP12Cert(cert *File, password *Secret) *Notorize {
	m.P12Cert = cert
	m.P12CertPassword = password

	return m
}

// WithNotoryKey sets the notory key and issuer
func (m *Notorize) WithNotoryKey(key *File, keyID, issuer string) *Notorize {
	m.NotoryKey = key
	m.NotoryKeyID = keyID
	m.NotoryIssuer = issuer

	return m
}

// Notarize notarizes a file using the notory key
func (m *Notorize) Notarize(ctx context.Context, file *File) error {
	return nil
}
