# A Go library for interacting with ACME (Automated Certificate Management Environment) servers, primarily designed for obtaining and managing SSL/TLS certificates from Let's Encrypt.

## Features

- Create and manage ACME accounts

- Handle domain validation challenges (HTTP-01 and DNS-01)

- Request and obtain SSL/TLS certificates

- Revoke certificates


## Installation 

```
go get github.com/disbeliefff/acme-lib
```

## Quick start

Here's a simple example of how to obtain a certificate:

```go
package main

import (
    "context"
    "log"
    
    "github.com/disbeliefff/acme-lib/acmeclient"
    "go.uber.org/zap"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Create client configuration
    config := acmeclient.Config{
        LEdir:  "https://acme-staging-v02.api.letsencrypt.org/directory", // Use staging URL for testing
        Logger: logger,
    }

    // Create new ACME client
    client, err := acmeclient.New(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Create or retrieve account
    ctx := context.Background()
    account, accountKey, err := client.CreateAccount(ctx, "admin@example.com")
    if err != nil {
        log.Fatalf("Failed to create account: %v", err)
    }

    // Use the client...
}
```

## Detailed Usage

1. Creating an Account

```go
account, accountKey, err := client.CreateAccount(ctx, "admin@example.com")
if err != nil {
    // Handle error
}
```

2. Handling Domain Validation

```go
// Request a challenge
challenge, err := client.GetChallenge(ctx, order, accountKey, acmeclient.ChallengeTypeHTTP)
if err != nil {
    // Handle error
}

// Generate challenge response
keyAuth, err := client.RespondChallenge(challenge, "example.com", accountKey)
if err != nil {
    // Handle error
}

// Set up your HTTP server to serve keyAuth at the required path
// or update your DNS records for DNS challenge
```

3. Finalizing Certificate Order

```go
// Assuming you have a CSR in PEM format
certs, certURL, err := client.FinalizeOrderWithCert(ctx, acmeClient, order, csrPem, true)
if err != nil {
    // Handle error
}

// Save your certificates
```
4. Revoking a Certificate

```go
err := client.RevokeCertificate(ctx, certPem, accountKey, 0)
if err != nil {
    // Handle error
}
```