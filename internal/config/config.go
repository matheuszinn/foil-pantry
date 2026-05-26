package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ValidatableConfig interface {
	Validate() error
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
}

func (d Database) Validate() error {
	if d.Host == "" {
		return errors.New("database.host is required")
	}
	if d.Port <= 0 {
		return errors.New("database.port must be greater than 0")
	}
	if d.User == "" {
		return errors.New("database.user is required")
	}
	if d.DbName == "" {
		return errors.New("database.dbname is required")
	}

	return nil
}

type Telegram struct {
	ApiID       int    `yaml:"api_id"`
	ApiHash     string `yaml:"api_hash"`
	Phone       string `yaml:"phone"`
	SessionPath string `yaml:"session_path"`
	SearchBot   string `yaml:"search_bot"`
}

func (t Telegram) Validate() error {
	if t.ApiID <= 0 {
		return errors.New("telegram.api_id must be greater than 0")
	}
	if t.ApiHash == "" {
		return errors.New("telegram.api_hash is required")
	}
	if t.Phone == "" {
		return errors.New("telegram.phone is required")
	}
	if t.SessionPath == "" {
		return errors.New("telegram.session_path is required")
	}
	if t.SearchBot == "" {
		return errors.New("telegram.search_bot is required")
	}

	return nil
}

type MinIO struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
}

func (m MinIO) Validate() error {
	if m.Endpoint == "" {
		return errors.New("minio.endpoint is required")
	}
	if m.AccessKey == "" {
		return errors.New("minio.access_key is required")
	}
	if m.SecretKey == "" {
		return errors.New("minio.secret_key is required")
	}
	if m.Bucket == "" {
		return errors.New("minio.bucket is required")
	}

	return nil
}

type Config struct {
	Database Database `yaml:"database"`
	Telegram Telegram `yaml:"telegram"`
	MinIO    MinIO    `yaml:"minio"`
}

func (c *Config) Validate() error {
	validatables := []ValidatableConfig{
		c.Database,
		c.MinIO,
		c.Telegram,
	}

	for _, v := range validatables {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {

	var cfg Config

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, &cfg)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
