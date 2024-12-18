package serv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"

	"github.com/disbeliefff/acme-lib/pkg/acme"
	"github.com/go-acme/lego/v4/certificate"
)

var processor *acme.ACMEProcessor

func Setup() (*http.Server, error) {
	var err error
	email := "sergey@me.com"
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Инициализация ACMEProcessor
	processor, err = acme.NewACMEProcessor(logger, email)
	if err != nil {
		fmt.Errorf("failed to initialize ACMEProcessor: %w", err)
	}

	// Настройка маршрутов
	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/validate", handleValidate)

	// Запуск сервера
	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Errorf("server failed to start: %w", err)
	}

	return nil, nil
}

type ChallengeResponse struct {
	ChallengeType string `json:"challengeType"`
	Domain        string `json:"domain"`
	Token         string `json:"token"`
	KeyAuth       string `json:"keyAuth"`
	ChallengeURL  string `json:"challengeURL"`
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	challengeType := r.URL.Query().Get("type")
	domain := r.URL.Query().Get("domain")

	if challengeType == "" || domain == "" {
		http.Error(w, "Missing query parameters: type or domain", http.StatusBadRequest)
		return
	}

	var err error
	switch challengeType {
	case "http-01":
		err = processor.SetHTTPChallengeProvider()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to set HTTP-01 provider: %v", err), http.StatusInternalServerError)
			return
		}
	case "dns-01":
		err = processor.SetDNSChallengeProvider()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to set DNS-01 provider: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unsupported challenge type", http.StatusBadRequest)
		return
	}

	certRequest := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	// Генерация сертификата и получения вызовов
	certificates, err := processor.GetClient().Certificate.Obtain(certRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to obtain certificate: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем сертификаты и вызовы
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certificates)
}

// Валидация challenge
func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request map[string]string
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	challengeURL := request["challengeURL"]
	if challengeURL == "" {
		http.Error(w, "Missing challenge URL", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Challenge validated successfully"))
}
