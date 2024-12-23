package acmeclient

import (
	"context"
	"crypto/ecdsa"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/disbeliefff/acme-lib/pkg/lib/logging"
	lib "github.com/disbeliefff/acme-lib/pkg/lib/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
)

// Challenge types supported by the ACME client
const (
	ChallengeTypeHTTP = "http-01"
	ChallengeTypeDNS  = "dns-01"
)

// Config represents the configuration for the ACME client
type Config struct {
	// LEdir is the Let's Encrypt directory URL
	LEdir string
	// Logger is the zap logger instance for logging
	Logger logging.Logger
}

// Client represents an ACME client instance that handles certificate operations
type Client struct {
	challenges sync.Map
	Config     Config
}

// New creates a new instance of the ACME client with the provided configuration
//
// Parameters:
//   - cfg: Configuration for the ACME client
//
// Returns:
//   - *Client: New ACME client instance
//   - error: Error if configuration is invalid
func New(cfg Config) (*Client, error) {
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	return &Client{
		Config: cfg,
	}, nil
}

// NewACMEClient creates a new ACME client instance with the provided private key
//
// Parameters:
//   - key: ECDSA private key for the ACME client
//
// Returns:
//   - *acme.Client: New ACME client instance
func (c *Client) NewACMEClient(key *ecdsa.PrivateKey) *acme.Client {
	c.Config.Logger.Info("creating new ACME client", zap.String("LEdir", c.Config.LEdir))
	return &acme.Client{
		Key:          key,
		DirectoryURL: c.Config.LEdir,
	}
}

// CreateAccount creates a new ACME account or retrieves an existing one
//
// Parameters:
//   - ctx: Context for the operation
//   - contact: Contact email for the account
//
// Returns:
//   - *acme.Account: Created or retrieved account
//   - *ecdsa.PrivateKey: Generated private key for the account
//   - error: Error if account creation fails
func (c *Client) CreateAccount(ctx context.Context, contact string) (*acme.Account, *ecdsa.PrivateKey, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c.Config.Logger.Info("creating new ACME account", zap.String("contact", contact))

	accountKey, err := lib.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("generate account key: %w", err)
	}

	client := c.NewACMEClient(accountKey)
	account := &acme.Account{Contact: []string{contact}}

	acc, err := client.Register(ctx, account, func(tosURL string) bool {
		c.Config.Logger.Info("accepting terms of service", zap.String("tosURL", tosURL))
		return true
	})

	if err != nil {
		if strings.Contains(err.Error(), "account already exists") {
			c.Config.Logger.Info("account already exists, retrieving", zap.String("contact", contact))
			acc, err = client.GetReg(ctx, "")
			if err != nil {
				return nil, nil, fmt.Errorf("get existing account: %w", err)
			}
			return acc, accountKey, nil
		}
		return nil, nil, fmt.Errorf("register account: %w", err)
	}

	c.Config.Logger.Info("account created successfully", zap.String("contact", acc.Contact[0]))
	return acc, accountKey, nil
}

// handleChallenge processes different types of ACME challenges
//
// Parameters:
//   - client: ACME client instance
//   - challenge: Challenge to handle
//   - token: Challenge token
//
// Returns:
//   - string: Challenge response
//   - error: Error if challenge handling fails
func (c *Client) handleChallenge(client *acme.Client, challenge *acme.Challenge, token string) (string, error) {
	switch challenge.Type {
	case ChallengeTypeHTTP:
		return client.HTTP01ChallengeResponse(token)
	case ChallengeTypeDNS:
		return client.DNS01ChallengeRecord(token)
	default:
		return "", fmt.Errorf("unsupported challenge type: %s", challenge.Type)
	}
}

// RespondChallenge generates and stores the response for an ACME challenge
//
// Parameters:
//   - challenge: ACME challenge to respond to
//   - domain: Domain name for the challenge
//   - accountKey: Account private key
//
// Returns:
//   - string: Challenge response
//   - error: Error if response generation fails
func (c *Client) RespondChallenge(challenge *acme.Challenge, domain string, accountKey *ecdsa.PrivateKey) (string, error) {
	c.Config.Logger.Info("responding to challenge",
		zap.String("domain", domain),
		zap.String("challengeType", challenge.Type))

	client := c.NewACMEClient(accountKey)
	keyAuth, err := c.handleChallenge(client, challenge, challenge.Token)
	if err != nil {
		return "", fmt.Errorf("handle challenge: %w", err)
	}

	c.challenges.Store(domain, challenge)
	return keyAuth, nil
}

// GetChallenge retrieves a specific type of challenge from an ACME order
//
// Parameters:
//   - order: ACME order containing authorizations
//   - accountKey: Account private key
//   - challengeType: Type of challenge to retrieve (http-01 or dns-01)
//
// Returns:
//   - *acme.Challenge: Retrieved challenge
//   - error: Error if challenge retrieval fails
func (c *Client) GetChallenge(ctx context.Context, order *acme.Order, accountKey *ecdsa.PrivateKey, challengeType string) (*acme.Challenge, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c.Config.Logger.Info("retrieving challenge",
		zap.String("challengeType", challengeType))

	client := c.NewACMEClient(accountKey)

	for _, authz := range order.AuthzURLs {
		authorization, err := client.GetAuthorization(ctx, authz)
		if err != nil {
			c.Config.Logger.Error("failed to fetch authorization",
				zap.String("authURL", authz),
				zap.Error(err))
			return nil, fmt.Errorf("fetch authorization %s: %w", authz, err)
		}

		for _, challenge := range authorization.Challenges {
			if challenge.Type == challengeType {
				c.Config.Logger.Info("challenge found",
					zap.String("challengeType", challengeType),
					zap.String("authURL", authz))
				return challenge, nil
			}
		}
	}

	c.Config.Logger.Error("no matching challenge found",
		zap.String("challengeType", challengeType))
	return nil, fmt.Errorf("no %s challenge found", challengeType)
}

// FinalizeOrderWithCert finalizes an ACME order with a Certificate Signing Request
//
// Parameters:
//   - ctx: Context for the operation
//   - client: ACME client instance
//   - order: ACME order to finalize
//   - csrPem: PEM-encoded Certificate Signing Request
//   - bundle: Whether to request the certificate bundle
//
// Returns:
//   - [][]byte: Certificate chain
//   - string: Certificate URL
//   - error: Error if finalization fails
func (c *Client) FinalizeOrderWithCert(ctx context.Context, client *acme.Client, order *acme.Order, csrPem []byte, bundle bool) ([][]byte, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c.Config.Logger.Info("finalizing order",
		zap.String("finalizeURL", order.FinalizeURL))

	csrBlock, _ := pem.Decode(csrPem)
	if csrBlock == nil || csrBlock.Type != "CERTIFICATE REQUEST" {
		c.Config.Logger.Error("invalid CSR format")
		return nil, "", errors.New("invalid CSR format")
	}

	certs, certURL, err := client.CreateOrderCert(ctx, order.FinalizeURL, csrBlock.Bytes, bundle)
	if err != nil {
		c.Config.Logger.Error("failed to finalize order",
			zap.Error(err))
		return nil, "", fmt.Errorf("create order certificate: %w", err)
	}

	c.Config.Logger.Info("order finalized successfully",
		zap.String("certURL", certURL))
	return certs, certURL, nil
}

// RevokeCertificate revokes a previously issued certificate
//
// Parameters:
//   - ctx: Context for the operation
//   - certPem: PEM-encoded certificate to revoke
//   - accountKey: Account private key
//   - reasonCode: Revocation reason code
//
// Returns:
//   - error: Error if revocation fails
func (c *Client) RevokeCertificate(ctx context.Context, certPem []byte, accountKey *ecdsa.PrivateKey, reasonCode int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	c.Config.Logger.Info("revoking certificate")

	block, _ := pem.Decode(certPem)
	if block == nil {
		return errors.New("invalid certificate format")
	}

	client := c.NewACMEClient(accountKey)
	if err := client.RevokeCert(ctx, accountKey, block.Bytes, 0); err != nil {
		return fmt.Errorf("revoke certificate: %w", err)
	}

	c.Config.Logger.Info("certificate revoked successfully")
	return nil
}
