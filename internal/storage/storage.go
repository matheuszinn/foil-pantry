package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageEngine interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error
}

type MinIOStorage struct {
	client *minio.Client
	bucket string
}

var _ StorageEngine = (*MinIOStorage)(nil)

func NewMinIOStorage(cfg *config.Config) (*MinIOStorage, error) {
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	return &MinIOStorage{
		client: minioClient,
		bucket: cfg.MinIO.Bucket,
	}, nil
}

func (m *MinIOStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	_, err = m.client.PutObject(ctx, m.bucket, objectName, reader, objectSize, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}
