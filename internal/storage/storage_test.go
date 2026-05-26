package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockStorage is a mock implementation of StorageEngine.
type MockStorage struct {
	UploadFunc func(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error
}

// UploadFile executes the mocked UploadFunc if defined.
func (m *MockStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, objectName, reader, objectSize)
	}
	return nil
}

// Ensure MockStorage implements StorageEngine.
var _ StorageEngine = (*MockStorage)(nil)

func TestInterfaceImplementation(t *testing.T) {
	// Verify that a MockStorage implements StorageEngine
	var _ StorageEngine = &MockStorage{}

	// Verify that MinIOStorage implements StorageEngine
	var _ StorageEngine = &MinIOStorage{}
}

func TestMockStorage_UploadFile(t *testing.T) {
	mockCalled := false
	mock := &MockStorage{
		UploadFunc: func(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
			mockCalled = true
			assert.Equal(t, "test-object.torrent", objectName)
			assert.Equal(t, int64(12), objectSize)
			content, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, "test content", string(content))
			return nil
		},
	}

	reader := strings.NewReader("test content")
	err := mock.UploadFile(context.Background(), "test-object.torrent", reader, 12)
	assert.NoError(t, err)
	assert.True(t, mockCalled)
}
