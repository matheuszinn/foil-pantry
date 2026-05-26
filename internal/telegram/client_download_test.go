package telegram

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadFile_InvalidFileIDs(t *testing.T) {
	wrapper := &ClientWrapper{} // empty client wrapper is enough to test initial parsing logic
	ctx := context.Background()

	t.Run("Empty file ID", func(t *testing.T) {
		err := wrapper.DownloadFile(ctx, "", io.Discard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file ID cannot be empty")
	})

	t.Run("Invalid parts count", func(t *testing.T) {
		err := wrapper.DownloadFile(ctx, "part1:part2", io.Discard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file ID format")
	})

	t.Run("Invalid document ID", func(t *testing.T) {
		err := wrapper.DownloadFile(ctx, "invalid:12345:aabbcc", io.Discard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document ID in file ID")
	})

	t.Run("Invalid access hash", func(t *testing.T) {
		err := wrapper.DownloadFile(ctx, "12345:invalid:aabbcc", io.Discard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid access hash in file ID")
	})

	t.Run("Invalid file reference hex", func(t *testing.T) {
		err := wrapper.DownloadFile(ctx, "12345:67890:invalid_hex", io.Discard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file reference hex in file ID")
	})
}
