package db

import (
	"time"
)

type DownloadStatus string

const (
	StatusPending       DownloadStatus = "PENDING"
	StatusMetadataReady DownloadStatus = "METADATA_READY"
	StatusFailed        DownloadStatus = "FAILED"
)

type Download struct {
	ID             uint           `gorm:"primaryKey;autoIncrement;column:id"`
	TitleName      string         `gorm:"column:title_name;type:varchar(255);not null"`
	TelegramMsgID  int            `gorm:"column:telegram_msg_id;not null"`
	TelegramFileID string         `gorm:"column:telegram_file_id;type:varchar(255);not null"`
	StoragePath    string         `gorm:"column:storage_path;type:varchar(255)"`
	Status         DownloadStatus `gorm:"column:status;type:varchar(50);default:'PENDING';not null"`
	SizeBytes      int64          `gorm:"column:size_bytes;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (Download) TableName() string {
	return "downloads"
}
