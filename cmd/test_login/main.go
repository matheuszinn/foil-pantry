package main

import (
	"context"
	"fmt"
	"log"

	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/telegram"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config.yaml: %v\nMake sure config.yaml exists and is valid. See config.yaml.example for structure.", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	wrapper, err := telegram.NewClientWrapper(cfg, logger)
	if err != nil {
		log.Fatalf("failed to initialize client wrapper: %v", err)
	}

	ctx := context.Background()

	err = wrapper.Run(ctx, func(ctx context.Context) error {
		fmt.Println("\n🎉 LOGGED IN SUCCESSFULLY TO TELEGRAM! 🎉")
		return nil
	})
	if err != nil {
		log.Fatalf("login execution failed: %v", err)
	}
}
