package main

import (
	"context"
	"fmt"
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
func (m *Notorize) SignAndNotorize(ctx context.Context, file *File) (*File, error) {
	ctr := m.getContainer().
		WithEnvVariable("QUILL_SIGN_P12", "/cert.p12").
		WithSecretVariable("QUILL_SIGN_PASSWORD", m.P12CertPassword).
		WithMountedFile("/cert.p12", m.P12Cert).
		WithEnvVariable("QUILL_NOTARY_KEY", "/key.p8").
		WithEnvVariable("QUILL_NOTARY_KEY_ID", m.NotoryKeyID).
		WithEnvVariable("QUILL_NOTARY_ISSUER", m.NotoryIssuer).
		WithEnvVariable("QUILL_LOG_QUIET", "false").
		WithEnvVariable("QUILL_LOG_FILE", "/quill.log").
		WithMountedFile("/key.p8", m.NotoryKey).
		WithMountedFile("/binary", file).
		WithExec([]string{"quill", "sign-and-notarize", "-vv", "/binary"})

	log := ctr.File("/quill.log")
	d, err := log.Contents(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println(d)

	return ctr.File("/binary"), nil
}

func (m *Notorize) TestNotorize(ctx context.Context, file *File, cert *File, password *Secret, key *File, keyId, keyIssuer string) (*File, error) {
	m.WithP12Cert(cert, password)
	m.WithNotoryKey(key, keyId, keyIssuer)

	return m.SignAndNotorize(ctx, file)
}

// +private
func (m *Notorize) getContainer() *Container {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "curl"}).
		WithExec([]string{"curl", "-L", "-o", "install.sh", "https://raw.githubusercontent.com/anchore/quill/main/install.sh"}).
		WithExec([]string{"chmod", "+x", "./install.sh"}).
		WithExec([]string{"./install.sh", "-b", "/usr/bin"})
}
