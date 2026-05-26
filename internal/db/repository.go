package db

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines the interface for database operations on the Download model.
type Repository interface {
	Create(ctx context.Context, download *Download) error
	UpdateStatusAndPath(ctx context.Context, id uint, status DownloadStatus, storagePath string) error
	GetByID(ctx context.Context, id uint) (*Download, error)
}

// SQLRepository implements Repository using GORM.
type SQLRepository struct {
	db *gorm.DB
}

// NewSQLRepository creates a new repository instance.
func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// Create inserts a new download record into the database.
func (r *SQLRepository) Create(ctx context.Context, download *Download) error {
	if download == nil {
		return errors.New("cannot create nil download record")
	}

	// Validate status is one of the allowed types
	if err := validateStatus(download.Status); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(download).Error; err != nil {
		return fmt.Errorf("failed to create download record: %w", err)
	}

	return nil
}

// UpdateStatusAndPath updates status and storage_path for a download.
func (r *SQLRepository) UpdateStatusAndPath(ctx context.Context, id uint, status DownloadStatus, storagePath string) error {
	if err := validateStatus(status); err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Model(&Download{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"storage_path": storagePath,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update download record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("download record with ID %d not found", id)
	}

	return nil
}

// GetByID retrieves a download record by its ID.
func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*Download, error) {
	var download Download
	if err := r.db.WithContext(ctx).First(&download, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("download record with ID %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to fetch download record: %w", err)
	}
	return &download, nil
}

// validateStatus checks that the status is valid.
func validateStatus(status DownloadStatus) error {
	switch status {
	case StatusPending, StatusMetadataReady, StatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid download status: %s", status)
	}
}
