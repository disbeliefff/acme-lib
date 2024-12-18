package providers_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	providers "github.com/disbeliefff/acme-lib/pkg/porviders"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thedevsaddam/gojsonq/v2"
)

func TestCustomDNSProvider_Singleton(t *testing.T) {
	provider1 := providers.GetInstance()
	provider2 := providers.GetInstance()

	assert.Equal(t, provider1, provider2, "Singleton instances should be the same")
}

func TestCustomDNSProvider_Present(t *testing.T) {
	provider := providers.GetInstance()
	domain := "primer.com"
	token := "test-token"

	// Simulate key authorization
	keyAuth := dns01.GetChallengeInfo(domain, token).Value

	err := provider.Present(domain, token, keyAuth)
	require.NoError(t, err, "Present method should not return an error")

	// Retrieve challenge from ready channel
	challenge, err := provider.GetChallenge()
	require.NoError(t, err, "Should be able to retrieve challenge")

	assert.Equal(t, domain, challenge.Identifier, "Challenge identifier should match domain")
	assert.Equal(t, "dns-01", challenge.Type, "Challenge type should be dns-01")
	assert.False(t, challenge.Verified, "Challenge should not be verified initially")
}

func TestCustomDNSProvider_CleanUp(t *testing.T) {
	provider := providers.GetInstance()
	domain := "cleanup-test.com"
	token := "cleanup-token"

	// Simulate key authorization
	keyAuth := dns01.GetChallengeInfo(domain, token).Value

	// First, present the challenge
	err := provider.Present(domain, token, keyAuth)
	require.NoError(t, err, "Present method should not return an error")

	// Retrieve the challenge first
	initialChallenge, err := provider.GetChallenge()
	require.NoError(t, err, "Should be able to retrieve initial challenge")

	// Verify the challenge details
	assert.Equal(t, domain, initialChallenge.Identifier, "Challenge identifier should match")

	// Then clean it up
	err = provider.CleanUp(domain, token, keyAuth)
	require.NoError(t, err, "CleanUp method should not return an error")

	// Optionally, add a method to check if a challenge exists
	exists := provider.ChallengeExists(domain, keyAuth)
	assert.False(t, exists, "Challenge should be removed after cleanup")
}
func TestCustomDNSProvider_MultiplePresent(t *testing.T) {
	provider := providers.GetInstance()
	domains := []string{
		"test1.com",
		"test2.com",
		"test3.com",
	}

	for _, domain := range domains {
		token := "token-" + domain
		keyAuth := dns01.GetChallengeInfo(domain, token).Value

		err := provider.Present(domain, token, keyAuth)
		require.NoError(t, err, "Present method should not return an error")
	}

	// Retrieve challenges
	for range domains {
		challenge, err := provider.GetChallenge()
		require.NoError(t, err, "Should be able to retrieve challenge")
		assert.Contains(t, []string{"test1.com", "test2.com", "test3.com"}, challenge.Identifier)
	}
}

func TestCustomDNSProvider_ConcurrentAccess(t *testing.T) {
	provider := providers.GetInstance()

	provider.ClearChallenges()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			domain := fmt.Sprintf("concurrent-test-%d.com", idx)
			token := fmt.Sprintf("token-%d", idx)
			keyAuth := dns01.GetChallengeInfo(domain, token).Value

			err := provider.Present(domain, token, keyAuth)
			assert.NoError(t, err, "Concurrent Present should not fail")
		}(i)
	}

	wg.Wait()

	// Проверяем, что все challenge были добавлены
	challengeCount := 0
	for {
		_, err := provider.GetChallenge()
		if err != nil {
			break
		}
		challengeCount++
	}

	assert.Equal(t, 100, challengeCount, "Should have 100 unique challenges")
}

func TestCustomDNSProvider_NoAvailableChallenges(t *testing.T) {
	provider := providers.GetInstance()

	// Очищаем все существующие challenge
	provider.ClearChallenges()

	// Try to get challenge when none are available
	_, err := provider.GetChallenge()
	assert.Error(t, err, "Should return error when no challenges are available")
}

func TestGetChallengesAsJSON_Structure(t *testing.T) {
	provider := providers.GetInstance()

	domain := "primer.com"
	keyAuth := "testKeyAuth"
	token := "testToken"

	err := provider.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	jsonOutput, err := provider.GetChallengesAsJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonOutput)

	var challenges []map[string]interface{}
	err = json.Unmarshal([]byte(jsonOutput), &challenges)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(challenges), 1, "There should be at least one challenge")
	for _, challenge := range challenges {
		assert.Contains(t, challenge, "type")
		assert.Contains(t, challenge, "identifier")
		assert.Contains(t, challenge, "content")
		assert.Contains(t, challenge, "verified")

		// Проверяем типы полей
		assert.IsType(t, "", challenge["type"])
		assert.IsType(t, "", challenge["identifier"])
		assert.IsType(t, "", challenge["content"])
		assert.IsType(t, false, challenge["verified"])
	}
}

func TestGetChallengesAsJSON_UsingGojsonq(t *testing.T) {
	provider := providers.GetInstance()

	domain := "primer.com"
	keyAuth := "testKeyAuth"
	token := "testToken"

	err := provider.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	jsonOutput, err := provider.GetChallengesAsJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonOutput)

	fmt.Println("JSON Output:", jsonOutput)

	jq := gojsonq.New().FromString(jsonOutput)

	identifier := jq.From("[0].identifier").Get()
	assert.NotNil(t, identifier, "Identifier should not be nil")
	assert.Equal(t, "primer.com", identifier)

	challengeType := jq.Reset().From("[0].type").Get()
	assert.NotNil(t, challengeType, "Type should not be nil")
	assert.Equal(t, "dns-01", challengeType)
}
