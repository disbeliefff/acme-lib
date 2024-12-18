package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/http01"
)

type HTTPProvider struct {
	tokenChan chan string
	domains   map[string]*ChallengeInfo
	mutex     sync.Mutex
}

type ChallengeInfo struct {
	Domain        string
	Token         string
	KeyAuth       string
	ChallengePath string
}

func NewHTTPProvider() (*HTTPProvider, error) {
	return &HTTPProvider{
		tokenChan: make(chan string, 1),
		domains:   make(map[string]*ChallengeInfo),
		mutex:     sync.Mutex{},
	}, nil
}

func (w *HTTPProvider) Present(domain, token, keyAuth string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Проверяем, был ли уже сгенерирован challenge для этого домена
	if challenge, ok := w.domains[domain]; ok {
		// Возвращаем ранее сгенерированный challenge
		fmt.Printf("Challenge already generated for domain: %s\n", domain)
		fmt.Printf("Challenge Info: %+v\n", challenge)
		return nil
	}

	// Генерация информации о вызове
	challengeInfo := &ChallengeInfo{
		Domain:        domain,
		Token:         token,
		KeyAuth:       keyAuth,
		ChallengePath: http01.ChallengePath(token),
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

	// Добавляем информацию о challenge в карту
	w.domains[domain] = challengeInfo

	// Ждем, пока пользователь вставит токен
	select {
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("timed out waiting for user to insert token")
	case keyAuth = <-w.tokenChan:
		fmt.Printf("Token received: %s\n", keyAuth)
	}

	return nil
}

func (w *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	// Ничего не делаем, оставляем файл с ключом на месте
	return nil
}

func (w *HTTPProvider) ProvideKeyAuth(keyAuth string) {
	w.tokenChan <- keyAuth
}
