package app

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/db"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/storage"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/telegram"
	"github.com/matheuszin/telegram-tinfoil-downloader/pkg/switchutil"
	"go.uber.org/zap"
)

// Run parses the configuration, establishes connections, executes the search, and manages the interactive selection and download.
func Run(ctx context.Context, configPath string, query string, verbose bool) error {
	var logger *zap.Logger
	var err error
	if verbose {
		logger, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		logger, err = cfg.Build()
	}
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Signal channel for SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case sig := <-sigChan:
			logger.Info("Received termination signal, initiating clean shutdown...", zap.String("signal", sig.String()))
			cancel()
		case <-ctx.Done():
		}
	}()

	logger.Info("Loading configuration...", zap.String("path", configPath))
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Connecting to MariaDB database...")
	dbConn, err := db.InitDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	sqlDB, err := dbConn.DB()
	if err == nil {
		defer func() {
			logger.Info("Closing database connection pool...")
			if err := sqlDB.Close(); err != nil {
				logger.Error("Error closing database connection pool", zap.Error(err))
			}
		}()
	}

	repo := db.NewSQLRepository(dbConn)

	logger.Info("Initializing MinIO storage engine...")
	storageEngine, err := storage.NewMinIOStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage engine: %w", err)
	}

	logger.Info("Initializing Telegram client...")
	clientWrapper, err := telegram.NewClientWrapper(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize telegram client: %w", err)
	}

	// Run the Telegram client execution loop
	err = clientWrapper.Run(ctx, func(ctx context.Context) error {
		logger.Info("Executing search on Telegram...", zap.String("query", query))
		var results []telegram.SearchResult
		err = retry(ctx, 3, 2*time.Second, logger, "telegram search", func() error {
			var err error
			results, err = clientWrapper.Search(ctx, query)
			return err
		})
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Printf("\nNo results found for query: %q\n", query)
			return nil
		}

		fmt.Println("\n==================================================")
		fmt.Println("              TELEGRAM DOWNLOADER")
		fmt.Println("==================================================")
		fmt.Printf("Search query: %q\n", query)
		fmt.Printf("Found %d results:\n\n", len(results))
		for _, r := range results {
			sizeStr := formatSize(r.Size)
			cleanName := switchutil.CleanTitle(r.Name)
			fmt.Printf(" [%d] %s\n", r.ID, cleanName)
			fmt.Printf("     Size: %s | Date: %s\n\n", sizeStr, r.Date.Format("2006-01-02 15:04:05"))
		}
		fmt.Println("--------------------------------------------------")

		reader := bufio.NewReader(os.Stdin)
		var choice int
		for {
			fmt.Printf("Enter the ID of the file you want to download (1-%d), or 0 to exit: ", len(results))
			text, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			text = strings.TrimSpace(text)
			if text == "0" || text == "" {
				fmt.Println("Download cancelled.")
				return nil
			}
			_, err = fmt.Sscanf(text, "%d", &choice)
			if err == nil && choice >= 1 && choice <= len(results) {
				break
			}
			fmt.Println("Invalid input. Please enter a valid index.")
		}

		selected := results[choice-1]
		cleanedTitle := switchutil.CleanTitle(selected.Name)

		download := &db.Download{
			TitleName:      cleanedTitle,
			TelegramMsgID:  selected.MsgID,
			TelegramFileID: selected.FileID,
			StoragePath:    fmt.Sprintf("torrents/%s", selected.Name),
			Status:         db.StatusPending,
			SizeBytes:      selected.Size,
		}

		logger.Info("Registering download metadata in database...", zap.String("filename", selected.Name))
		if err := repo.Create(ctx, download); err != nil {
			return fmt.Errorf("failed to create download record in database: %w", err)
		}

		failDownload := func(reasonErr error) error {
			logger.Error("Operation failed, updating status to FAILED", zap.Error(reasonErr))
			if updateErr := repo.UpdateStatusAndPath(ctx, download.ID, db.StatusFailed, download.StoragePath); updateErr != nil {
				logger.Error("Failed to update status to FAILED in database", zap.Error(updateErr))
			}
			return reasonErr
		}

		logger.Info("Downloading torrent file from Telegram...", zap.String("fileID", selected.FileID))
		var buf bytes.Buffer
		err = retry(ctx, 3, 2*time.Second, logger, "telegram download", func() error {
			buf.Reset()
			return clientWrapper.DownloadFile(ctx, selected.FileID, &buf)
		})
		if err != nil {
			return failDownload(fmt.Errorf("failed to download file from Telegram: %w", err))
		}

		logger.Info("Uploading torrent file to MinIO...", zap.String("storagePath", download.StoragePath))
		readerStream := bytes.NewReader(buf.Bytes())
		err = retry(ctx, 3, 2*time.Second, logger, "minio upload", func() error {
			_, _ = readerStream.Seek(0, io.SeekStart)
			return storageEngine.UploadFile(ctx, download.StoragePath, readerStream, int64(buf.Len()))
		})
		if err != nil {
			return failDownload(fmt.Errorf("failed to upload torrent file to MinIO: %w", err))
		}

		logger.Info("Updating database status to METADATA_READY...")
		if err := repo.UpdateStatusAndPath(ctx, download.ID, db.StatusMetadataReady, download.StoragePath); err != nil {
			return fmt.Errorf("failed to update status to METADATA_READY in database: %w", err)
		}

		logger.Info("Torrent metadata successfully captured and persisted!")
		return nil
	})

	if err != nil {
		return fmt.Errorf("app execution error: %w", err)
	}

	return nil
}

// retry helper executing a given operation up to the specified number of attempts with a sleep delay in between.
func retry(ctx context.Context, attempts int, delay time.Duration, logger *zap.Logger, opName string, op func() error) error {
	var err error
	for i := 1; i <= attempts; i++ {
		if err = op(); err == nil {
			return nil
		}
		logger.Warn("Operation failed, retrying...",
			zap.String("operation", opName),
			zap.Int("attempt", i),
			zap.Int("max_attempts", attempts),
			zap.Error(err),
		)
		if i < attempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w", opName, attempts, err)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
