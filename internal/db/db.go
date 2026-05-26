package db

import (
	"fmt"

	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB initializes a connection to MariaDB and runs automigrations.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DbName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MariaDB database: %w", err)
	}

	if err := db.AutoMigrate(&Download{}); err != nil {
		return nil, fmt.Errorf("failed to run database automigration: %w", err)
	}

	return db, nil
}
