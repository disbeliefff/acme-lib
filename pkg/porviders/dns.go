package providers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

type DNSChallenge struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	Content    string `json:"content"`
	Verified   bool   `json:"verified"`
}

type CustomDNSProvider struct {
	challenges   []DNSChallenge
	mu           sync.RWMutex
	ready        chan DNSChallenge
	resolveDelay time.Duration // добавляем задержку
}

var (
	instance *CustomDNSProvider
	once     sync.Once
)

func GetInstance() *CustomDNSProvider {
	once.Do(func() {
		instance = &CustomDNSProvider{
			challenges:   []DNSChallenge{},
			ready:        make(chan DNSChallenge, 100),
			resolveDelay: 60 * time.Second,
		}
	})
	return instance
}

// Present registers the challenge and adds a delay before resolution
func (p *CustomDNSProvider) Present(domain, token, keyAuth string) error {
	challengeInfo := dns01.GetChallengeInfo(domain, keyAuth)
	txtValue := challengeInfo.Value

	challenge := DNSChallenge{
		Type:       "dns-01",
		Identifier: domain,
		Content:    txtValue,
		Verified:   false,
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for duplicate challenges
	for _, existing := range p.challenges {
		if existing.Identifier == domain && existing.Content == txtValue {
			return nil
		}
	}

	// Add the new challenge
	p.challenges = append(p.challenges, challenge)

	select {
	case p.ready <- challenge:
	default:
		fmt.Println("Challenge channel is full")
	}

	// Return JSON response immediately
	challengeJSON, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("could not marshal challenge to JSON: %w", err)
	}

	fmt.Printf("Challenge Info: %s\n", challengeJSON)

	// Add delay before resolution
	go func() {
		time.Sleep(p.resolveDelay)
		// Можно добавить дополнительную логику для обработки challenge после задержки
		fmt.Printf("Challenge for domain %s is now ready to resolve\n", domain)
	}()

	return nil
}

func (p *CustomDNSProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	challengeInfo := dns01.GetChallengeInfo(domain, keyAuth)
	txtValue := challengeInfo.Value

	var updatedChallenges []DNSChallenge
	for _, c := range p.challenges {
		if !(c.Identifier == domain && c.Content == txtValue) {
			updatedChallenges = append(updatedChallenges, c)
		}
	}

	p.challenges = updatedChallenges

	fmt.Printf("Cleanup for domain: %s, token: %s\n", domain, token)
	return nil
}

func (p *CustomDNSProvider) GetChallengesAsJSON() (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	challengesJSON, err := json.Marshal(p.challenges)
	if err != nil {
		return "", fmt.Errorf("could not marshal challenges to JSON: %w", err)
	}

	return string(challengesJSON), nil
}

func (p *CustomDNSProvider) GetChallenge() (DNSChallenge, error) {
	select {
	case challenge := <-p.ready:
		return challenge, nil
	default:
		return DNSChallenge{}, fmt.Errorf("no challenges available")
	}
}

func (p *CustomDNSProvider) ClearChallenges() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.challenges = []DNSChallenge{}

	for {
		select {
		case <-p.ready:
		default:
			return
		}
	}
}

func (p *CustomDNSProvider) ChallengeExists(domain, content string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, challenge := range p.challenges {
		if challenge.Identifier == domain && challenge.Content == content {
			return true
		}
	}
	return false
}
