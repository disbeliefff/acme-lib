package providers_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	providers "github.com/disbeliefff/acme-lib/pkg/porviders"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPProvider_Present(t *testing.T) {
	provider, err := providers.NewHTTPProvider()
	require.NoError(t, err)

	domain := "example.com"
	token := "testToken123"
	keyAuth := "testKeyAuth123"

	// Вызываем метод Present
	err = provider.Present(domain, token, keyAuth)
	require.NoError(t, err)

	// Проверяем, что файл с токеном создан
	tokenFilePath := filepath.Join(".", token)
	content, err := os.ReadFile(tokenFilePath)
	require.NoError(t, err)
	assert.Equal(t, keyAuth, string(content))

	// Убираем файл после теста
	defer os.Remove(tokenFilePath)
}

func TestHTTPProvider_CleanUp(t *testing.T) {
	provider, err := providers.NewHTTPProvider()
	require.NoError(t, err)

	domain := "example.com"
	token := "testToken123"
	keyAuth := "testKeyAuth123"

	// Создаем токен
	err = provider.Present(domain, token, keyAuth)
	require.NoError(t, err)

	// Проверяем, что файл существует
	tokenFilePath := filepath.Join(".", token)
	_, err = os.Stat(tokenFilePath)
	require.NoError(t, err)

	// Вызываем CleanUp
	err = provider.CleanUp(domain, token, keyAuth)
	require.NoError(t, err)

	// Проверяем, что файл удален
	_, err = os.Stat(tokenFilePath)
	assert.True(t, os.IsNotExist(err))
}

func TestHTTPProvider_GetChallengeInfoAsJSON(t *testing.T) {
	provider, err := providers.NewHTTPProvider()
	require.NoError(t, err)

	domain := "example.com"
	token := "testToken123"
	keyAuth := "testKeyAuth123"

	// Вызываем Present
	err = provider.Present(domain, token, keyAuth)
	require.NoError(t, err)

	// Генерируем challenge info в JSON
	challengeInfo := map[string]string{
		"domain":        domain,
		"token":         token,
		"keyAuth":       keyAuth,
		"challengePath": http01.ChallengePath(token),
	}

	expectedJSON, err := json.Marshal(challengeInfo)
	require.NoError(t, err)

	// Проверяем, что JSON корректный
	fmt.Println("Expected JSON:", string(expectedJSON))
}

func TestHTTPProvider_ConcurrentAccess(t *testing.T) {
	provider, err := providers.NewHTTPProvider()
	require.NoError(t, err)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			domain := fmt.Sprintf("concurrent-%d.com", idx)
			token := fmt.Sprintf("token-%d", idx)
			keyAuth := fmt.Sprintf("keyAuth-%d", idx)

			// Проверяем Present
			err := provider.Present(domain, token, keyAuth)
			assert.NoError(t, err)

			// Проверяем создание файла
			tokenFilePath := filepath.Join(".", token)
			_, err = os.Stat(tokenFilePath)
			assert.NoError(t, err)

			// Чистим после себя
			err = provider.CleanUp(domain, token, keyAuth)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}
