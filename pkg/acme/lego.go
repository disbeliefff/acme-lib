package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"log/slog"

	providers "github.com/disbeliefff/acme-lib/pkg/porviders"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type ACMEProcessor struct {
	Client *lego.Client
	lg     *slog.Logger
}

type MyUser struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}

func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

func GeneratePrivateKey() (crypto.PrivateKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return privKey, nil
}

func NewACMEProcessor(logger *slog.Logger, email string) (*ACMEProcessor, error) {
	logger.Info("Initializing ACMEProcessor...")

	logger.Info("Generating private key...")
	privKey, err := GeneratePrivateKey()
	if err != nil {
		logger.Error("Failed to generate private key", "error", err)
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	user := &MyUser{
		Email:        email,
		Registration: nil,
		Key:          privKey,
	}

	logger.Info("Configuring ACME client...")
	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryStaging

	logger.Info("Creating ACME client...")
	client, err := lego.NewClient(config)
	if err != nil {
		logger.Error("Failed to create ACME client", "error", err)
		return nil, fmt.Errorf("failed to create ACME client: %w", err)
	}

	logger.Info("Registering user with ACME server...")
	regOptions := registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	}
	reg, err := client.Registration.Register(regOptions)
	if err != nil {
		logger.Error("Failed to register user", "error", err)
		return nil, fmt.Errorf("failed to register user: %w", err)
	}
	user.Registration = reg

	logger.Info("ACMEProcessor initialized successfully.")
	return &ACMEProcessor{
		Client: client,
		lg:     logger,
	}, nil
}

func (c *ACMEProcessor) SetHTTPChallengeProvider() error {
	c.lg.Info("Setting HTTP-01 challenge provider...")
	httpProvider, err := providers.NewHTTPProvider()
	if err != nil {
		c.lg.Error("Failed to initialize HTTP provider", "error", err)
		return fmt.Errorf("failed to initialize HTTP provider: %w", err)
	}

	if err := c.Client.Challenge.SetHTTP01Provider(httpProvider); err != nil {
		c.lg.Error("Failed to set HTTP provider", "error", err)
		return fmt.Errorf("failed to set HTTP provider: %w", err)
	}

	c.lg.Info("HTTP-01 challenge provider set successfully.")
	return nil
}

func (c *ACMEProcessor) SetDNSChallengeProvider() error {
	c.lg.Info("Setting DNS-01 challenge provider...")
	dnsProvider := providers.GetInstance()

	if err := c.Client.Challenge.SetDNS01Provider(dnsProvider); err != nil {
		c.lg.Error("Failed to set DNS provider", "error", err)
		return fmt.Errorf("failed to set DNS provider: %w", err)
	}

	c.lg.Info("DNS-01 challenge provider set successfully.")
	return nil
}

func (c *ACMEProcessor) GetClient() *lego.Client {
	return c.Client
}

// Внутри pkg/acme/ACMEProcessor.go

// ChallengeInfo представляет информацию о вызове
type ChallengeInfo struct {
	Token   string
	KeyAuth string
	Path    string
}
