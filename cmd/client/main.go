package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yw/lipa-na-mpesa-go/internal/daraja"
)

func main() {
	if err := run(); err != nil {
		slog.Error("C2B command failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("load .env file: %w", err)
	}

	if len(os.Args) < 2 {
		return fmt.Errorf("usage: go run ./cmd/client [register|simulate]")
	}

	command := os.Args[1]

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	client := daraja.NewHTTPClient()
	config := daraja.ConfigFromEnv()

	accessToken, err := daraja.GetAccessToken(ctx, client, config)
	if err != nil {
		return fmt.Errorf("OAuth failed: %w", err)
	}

	slog.Debug("OAuth access token obtained")

	switch command {
	case "register":
		if err := daraja.RegisterC2BURLs(
			ctx, client, config, accessToken,
		); err != nil {
			return fmt.Errorf("C2B registration failed: %w", err)
		}

		slog.Info("C2B callback URLs registered")

	case "simulate":
		if err := daraja.SimulateC2B(
			ctx, client, config, accessToken,
		); err != nil {
			return fmt.Errorf("C2B simulation failed: %w", err)
		}

		slog.Info("C2B simulation request accepted")

	default:
		return fmt.Errorf("unknown command %q (use register or simulate)", command)
	}
	return nil
}
