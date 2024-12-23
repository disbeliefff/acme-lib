package acmeclient

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"sync"

	"github.com/disbeliefff/acme-lib/pkg/lib"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
)

type Client struct {
	challenges sync.Map
	LEdir      string
	lg         *zap.Logger
}

func New(LEdir string, lg *zap.Logger) *Client {
	return &Client{
		LEdir: LEdir,
		lg:    lg,
	}
}

func (c *Client) NewACMEClient(key *ecdsa.PrivateKey) *acme.Client {
	c.lg.Info("Creating new ACME client", zap.String("LEdir", c.LEdir))
	return &acme.Client{
		Key:          key,
		DirectoryURL: c.LEdir,
	}
}

func (c *Client) CreateAccount(contact string) (*acme.Account, *ecdsa.PrivateKey, error) {
	c.lg.Info("Creating new ACME account", zap.String("contact", contact))

	accountKey, err := lib.GenerateKey()
	if err != nil {
		c.lg.Error("Failed to generate account key", zap.Error(err))
		return nil, nil, err
	}

	client := c.NewACMEClient(accountKey)
	ctx := context.Background()

	account := &acme.Account{
		Contact: []string{contact},
	}

	acc, err := client.Register(ctx, account, func(tosURL string) bool {
		c.lg.Info("Registering account", zap.String("tosURL", tosURL))
		return true
	})
	if err != nil {
		if strings.Contains(err.Error(), "account already exists") {
			c.lg.Error("Account already exists for contact: %s", zap.String("contact", contact))
			acc, err = client.GetReg(ctx, "")
			if err != nil {
				c.lg.Error("Error retrieving existing account: %v", zap.Error(err))
				return nil, nil, err
			}
			return acc, accountKey, nil
		}
		c.lg.Error("Error creating account: %v", zap.Error(err))
		return nil, nil, err
	}

	c.lg.Error("Account created", zap.String("contact", acc.Contact[0]))
	return acc, accountKey, nil
}

func (c *Client) RequestOrder(domain string, accountKey *ecdsa.PrivateKey) (*acme.Order, error) {
	c.lg.Info("Requesting order for domain", zap.String("domain", domain))

	client := c.NewACMEClient(accountKey)

	ctx := context.Background()

	order, err := client.AuthorizeOrder(ctx, acme.DomainIDs(domain))
	if err != nil {
		c.lg.Error("Error authorizing order: %v", zap.Error(err))
		return nil, err
	}

	c.lg.Info("Order created", zap.String("domain", domain))
	return order, nil
}
