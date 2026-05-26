package config_test

import (
	"testing"

	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLocalConfig(t *testing.T) {

	expected := config.Config{
		Database: config.Database{
			Host:     "localhost",
			Port:     3306,
			User:     "user",
			Password: "password",
			DbName:   "database",
		},
		Telegram: config.Telegram{
			ApiID:       12345,
			ApiHash:     "your_api_hash_here",
			Phone:       "+5511999999999",
			SessionPath: "telegram.session",
			SearchBot:   "BotUsername",
		},
		MinIO: config.MinIO{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			Bucket:    "bucketname",
			UseSSL:    false,
		},
	}

	actual, err := config.LoadConfig("resources/config.yaml.example")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, expected, *actual)
}

func TestConfigValidation(t *testing.T) {
	baseConfig := config.Config{
		Database: config.Database{
			Host:   "localhost",
			Port:   3306,
			User:   "root",
			DbName: "test",
		},
		Telegram: config.Telegram{
			ApiID:       12345,
			ApiHash:     "hash",
			Phone:       "12345",
			SessionPath: "session",
			SearchBot:   "bot",
		},
		MinIO: config.MinIO{
			Endpoint:  "localhost:9000",
			AccessKey: "minio",
			SecretKey: "minio123",
			Bucket:    "bucket",
		},
	}

	t.Run("Valid configuration", func(t *testing.T) {
		cfg := baseConfig
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing Database.Host", func(t *testing.T) {
		cfg := baseConfig
		cfg.Database.Host = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.host is required")
	})

	t.Run("Invalid Database.Port", func(t *testing.T) {
		cfg := baseConfig
		cfg.Database.Port = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.port must be greater than 0")
	})

	t.Run("Missing Database.User", func(t *testing.T) {
		cfg := baseConfig
		cfg.Database.User = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.user is required")
	})

	t.Run("Missing Database.DbName", func(t *testing.T) {
		cfg := baseConfig
		cfg.Database.DbName = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.dbname is required")
	})

	t.Run("Invalid Telegram.ApiID", func(t *testing.T) {
		cfg := baseConfig
		cfg.Telegram.ApiID = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram.api_id must be greater than 0")
	})

	t.Run("Missing Telegram.ApiHash", func(t *testing.T) {
		cfg := baseConfig
		cfg.Telegram.ApiHash = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram.api_hash is required")
	})

	t.Run("Missing Telegram.Phone", func(t *testing.T) {
		cfg := baseConfig
		cfg.Telegram.Phone = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram.phone is required")
	})

	t.Run("Missing Telegram.SessionPath", func(t *testing.T) {
		cfg := baseConfig
		cfg.Telegram.SessionPath = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram.session_path is required")
	})

	t.Run("Missing Telegram.SearchBot", func(t *testing.T) {
		cfg := baseConfig
		cfg.Telegram.SearchBot = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram.search_bot is required")
	})

	t.Run("Missing MinIO.Endpoint", func(t *testing.T) {
		cfg := baseConfig
		cfg.MinIO.Endpoint = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minio.endpoint is required")
	})

	t.Run("Missing MinIO.AccessKey", func(t *testing.T) {
		cfg := baseConfig
		cfg.MinIO.AccessKey = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minio.access_key is required")
	})

	t.Run("Missing MinIO.SecretKey", func(t *testing.T) {
		cfg := baseConfig
		cfg.MinIO.SecretKey = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minio.secret_key is required")
	})

	t.Run("Missing MinIO.Bucket", func(t *testing.T) {
		cfg := baseConfig
		cfg.MinIO.Bucket = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minio.bucket is required")
	})
}
