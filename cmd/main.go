package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/disbeliefff/acme-lib/internal/serv"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server, err := serv.Setup()
	if err != nil {
		logger.Error("Failed to setup server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is running on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed to start", "error", err)
	}
}
