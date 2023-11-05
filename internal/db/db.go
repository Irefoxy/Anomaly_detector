package db

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type Config struct {
	DBUsername string `yaml:"DBUsername"`
	DBPassword string `yaml:"DBPassword"`
	DBName     string `yaml:"DBName"`
	DBHost     string `yaml:"DBHost"`
	DBPort     string `yaml:"DBPort"`
}

type Record struct {
	gorm.Model
	Uuid      string
	Frequency float64
	Timestamp int64
}

func Connect(configPath string) (*gorm.DB, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&Record{}); err != nil {
		return nil, err
	}
	return db, nil
}

func loadConfig(configPath string) (*Config, error) {
	var config Config

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
