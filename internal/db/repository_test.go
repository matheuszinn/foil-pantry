package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Download{})
	require.NoError(t, err)

	return db
}

func TestSQLRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		download := &Download{
			TitleName:      "Super Mario Odyssey",
			TelegramMsgID:  12345,
			TelegramFileID: "file_abc_123",
			SizeBytes:      1024 * 1024 * 100, // 100MB
			Status:         StatusPending,
		}

		err := repo.Create(ctx, download)
		require.NoError(t, err)
		assert.NotZero(t, download.ID)

		// Verify record is retrieved correctly
		fetched, err := repo.GetByID(ctx, download.ID)
		require.NoError(t, err)
		assert.Equal(t, "Super Mario Odyssey", fetched.TitleName)
		assert.Equal(t, 12345, fetched.TelegramMsgID)
		assert.Equal(t, "file_abc_123", fetched.TelegramFileID)
		assert.Equal(t, StatusPending, fetched.Status)
		assert.Equal(t, int64(1024*1024*100), fetched.SizeBytes)
	})

	t.Run("nil record error", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot create nil download record")
	})

	t.Run("invalid status validation error", func(t *testing.T) {
		download := &Download{
			TitleName:      "Zelda",
			TelegramMsgID:  999,
			TelegramFileID: "file_xyz",
			SizeBytes:      5000,
			Status:         DownloadStatus("COMPLETED"), // Invalid status
		}

		err := repo.Create(ctx, download)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid download status")
	})
}

func TestSQLRepository_UpdateStatusAndPath(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := context.Background()

	// Setup initial record
	download := &Download{
		TitleName:      "Metroid Dread",
		TelegramMsgID:  777,
		TelegramFileID: "file_metroid",
		SizeBytes:      8000,
		Status:         StatusPending,
	}
	err := repo.Create(ctx, download)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		storagePath := "torrents/metroid.torrent"
		err := repo.UpdateStatusAndPath(ctx, download.ID, StatusMetadataReady, storagePath)
		require.NoError(t, err)

		// Verify status and path updated
		fetched, err := repo.GetByID(ctx, download.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusMetadataReady, fetched.Status)
		assert.Equal(t, storagePath, fetched.StoragePath)
	})

	t.Run("invalid status update", func(t *testing.T) {
		err := repo.UpdateStatusAndPath(ctx, download.ID, DownloadStatus("UNKNOWN"), "path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid download status")
	})

	t.Run("non-existent record", func(t *testing.T) {
		err := repo.UpdateStatusAndPath(ctx, 9999, StatusFailed, "path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "download record with ID 9999 not found")
	})
}

func TestSQLRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := context.Background()

	t.Run("non-existent record returns error", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 9999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
