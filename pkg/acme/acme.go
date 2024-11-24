package acme

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"github.com/go-acme/lego/registration"
)

type ACME struct {
	Email        string
	agreeTerms   bool
	keyType      string
	privateKey   *rsa.PrivateKey
	registration *registration.Resource
}

func New(email string, agreeTerms bool, keyType string) (*ACME, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096) // using 4096 bit certs
	if err != nil {
		return nil, err
	}

	acme := &ACME{
		Email:        email,
		agreeTerms:   agreeTerms,
		keyType:      keyType,
		privateKey:   privateKey,
		registration: registration,
	}

	return acme, nil
}

func (a *ACME) ObtainCertificate(ctx context.Context, domains []string) (*tls.Certificate, error) {
	return nil, nil
}

func (a *ACME) RenewCertificate(ctx context.Context, domains []string) (*tls.Certificate, error) {

	return nil, nil
}
