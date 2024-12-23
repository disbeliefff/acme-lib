package main

import (
	"log"

	acmeclient "github.com/disbeliefff/acme-lib/pkg/acme-client"
	"github.com/disbeliefff/acme-lib/pkg/lib/logging"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer zapLogger.Sync()

	logger := logging.NewZapLogger(zapLogger)

	// Initialize ACME client with Let's Encrypt directory URL
	cfg := acmeclient.Config{
		LEdir:  "https://acme-v02.api.letsencrypt.org/directory",
		Logger: logger,
	}

	client, err := acmeclient.New(cfg)
	if err != nil {
		log.Fatalf("failed to create ACME client: %v", err)
	}

	client.Config.Logger.Info("ACME client initialized", "LEdir", cfg.LEdir)
}
