package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/challenge/http01"
)

type HTTPProvider struct{}

func NewHTTPProvider() (*HTTPProvider, error) {
	return &HTTPProvider{}, nil
}

func (w *HTTPProvider) Present(domain, token, keyAuth string) error {
	// Генерация информации о вызове
	challengeInfo := map[string]string{
		"domain":        domain,
		"token":         token,
		"keyAuth":       keyAuth,
		"challengePath": http01.ChallengePath(token),
	}

	challengeJSON, err := json.Marshal(challengeInfo)
	if err != nil {
		return fmt.Errorf("could not marshal challenge info to JSON: %w", err)
	}

	fmt.Printf("Challenge Info: %s\n", challengeJSON)

	tokenFilePath := token
	err = os.WriteFile(tokenFilePath, []byte(keyAuth), 0600)
	if err != nil {
		return fmt.Errorf("could not create token file: %w", err)
	}

	fmt.Printf("Token file created: %s\n", tokenFilePath)
	return nil
}

func (w *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	tokenFilePath := token
	err := os.Remove(tokenFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove token file: %w", err)
	}

	fmt.Printf("Token file removed: %s\n", tokenFilePath)
	return nil
}
