package db

import (
	"fmt"
	"time"

	"backend/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQL struct {
	DB *gorm.DB
}

func NewMySQL(cfg *config.Config) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	gcfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(mysql.Open(dsn), gcfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(&Auction{}, &Bid{}, &SyncCursor{}); err != nil {
		return nil, err
	}

	return &MySQL{DB: db}, nil
}
