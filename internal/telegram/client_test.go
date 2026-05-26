package telegram

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewClientWrapper_NilInputs(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Telegram: config.Telegram{
			ApiID:       12345,
			ApiHash:     "dummyhash",
			Phone:       "+1234567890",
			SessionPath: filepath.Join(t.TempDir(), "session.json"),
			SearchBot:   "dummy_bot",
		},
	}

	_, err := NewClientWrapper(nil, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")

	_, err = NewClientWrapper(cfg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger is nil")
}

func TestNewClientWrapper_CreateSessionDir(t *testing.T) {
	tempDir := t.TempDir()
	sessionPath := filepath.Join(tempDir, "non_existent_subdir", "session.json")

	cfg := &config.Config{
		Telegram: config.Telegram{
			ApiID:       12345,
			ApiHash:     "dummyhash",
			Phone:       "+1234567890",
			SessionPath: sessionPath,
			SearchBot:   "dummy_bot",
		},
	}

	logger := zap.NewNop()

	// Ensure the parent directory does not exist yet
	parentDir := filepath.Dir(sessionPath)
	_, err := os.Stat(parentDir)
	assert.True(t, os.IsNotExist(err), "parent directory should not exist yet")

	// Instantiate ClientWrapper
	wrapper, err := NewClientWrapper(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, wrapper)

	// Verify that the parent directory was successfully created
	_, err = os.Stat(parentDir)
	assert.NoError(t, err, "parent directory should have been created by NewClientWrapper")
}

func TestInteractiveAuthenticator_Phone(t *testing.T) {
	ctx := context.Background()

	t.Run("Default phone fallback", func(t *testing.T) {
		input := bytes.NewBufferString("\n") // user presses Enter
		auth := &interactiveAuthenticator{
			defaultPhone: "+1234567890",
			stdin:        input,
		}

		phone, err := auth.Phone(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "+1234567890", phone)
	})

	t.Run("Custom phone input", func(t *testing.T) {
		input := bytes.NewBufferString("+9876543210\n") // user enters a different number
		auth := &interactiveAuthenticator{
			defaultPhone: "+1234567890",
			stdin:        input,
		}

		phone, err := auth.Phone(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "+9876543210", phone)
	})
}

func TestInteractiveAuthenticator_Password(t *testing.T) {
	ctx := context.Background()
	input := bytes.NewBufferString("my_2fa_password\n")
	auth := &interactiveAuthenticator{
		stdin: input,
	}

	password, err := auth.Password(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "my_2fa_password", password)
}

func TestInteractiveAuthenticator_Code(t *testing.T) {
	ctx := context.Background()
	input := bytes.NewBufferString("12345\n")
	auth := &interactiveAuthenticator{
		stdin: input,
	}

	code, err := auth.Code(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, "12345", code)
}

func TestInteractiveAuthenticator_SignUp(t *testing.T) {
	ctx := context.Background()
	auth := &interactiveAuthenticator{}

	_, err := auth.SignUp(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sign up not supported")
}

func TestInteractiveAuthenticator_AcceptTermsOfService(t *testing.T) {
	ctx := context.Background()
	auth := &interactiveAuthenticator{}

	err := auth.AcceptTermsOfService(ctx, tg.HelpTermsOfService{})
	assert.NoError(t, err)
}
