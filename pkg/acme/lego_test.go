package acme_test

import (
	"os"
	"testing"

	"github.com/disbeliefff/acme-lib/pkg/acme"

	"log/slog"

	"github.com/go-acme/lego/v4/registration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewACMEProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	email := "yous@you.com"

	processor, err := acme.NewACMEProcessor(logger, email)
	require.NoError(t, err, "Failed to initialize ACMEProcessor")
	assert.NotNil(t, processor, "Processor should not be nil")
	assert.NotNil(t, processor.Client, "Client should not be nil")
}

func TestSetHTTPChallengeProvider(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	email := "yous@you.com"

	processor, err := acme.NewACMEProcessor(logger, email)
	require.NoError(t, err, "Failed to initialize ACMEProcessor")

	err = processor.SetHTTPChallengeProvider()
	assert.NoError(t, err, "Failed to set HTTP-01 challenge provider")
}

func TestSetDNSChallengeProvider(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	email := "yous@you.com"

	processor, err := acme.NewACMEProcessor(logger, email)
	require.NoError(t, err, "Failed to initialize ACMEProcessor")

	err = processor.SetDNSChallengeProvider()
	assert.NoError(t, err, "Failed to set DNS-01 challenge provider")
}

func TestUserImplementation(t *testing.T) {
	email := "yous@you.com"
	key, err := acme.GeneratePrivateKey()
	require.NoError(t, err, "Failed to generate private key")

	user := &acme.MyUser{
		Email:        email,
		Registration: &registration.Resource{},
		Key:          key,
	}

	assert.Equal(t, email, user.GetEmail(), "Email mismatch")
	assert.NotNil(t, user.GetPrivateKey(), "Private key should not be nil")
	assert.NotNil(t, user.GetRegistration(), "Registration resource should not be nil")
}
