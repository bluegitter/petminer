package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize 初始化数据库连接
func Initialize() error {
	// 确保data目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 数据库文件路径
	dbPath := "data/petminer.db"
	
	// 配置GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// SQLite连接池配置
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
	return nil
}

// Migrate 执行数据库迁移
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("Running database migrations...")

	// 自动迁移数据库表
	if err := DB.AutoMigrate(&DBPet{}, &DBEvent{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// Health 检查数据库健康状态
func Health() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// CleanupOldEvents 清理旧事件，保持数据库大小合理
func CleanupOldEvents(keepCount int) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 计算当前事件总数
	var totalCount int64
	if err := DB.Model(&DBEvent{}).Count(&totalCount).Error; err != nil {
		return err
	}

	if totalCount <= int64(keepCount) {
		return nil // 不需要清理
	}

	// 删除最老的事件
	deleteCount := totalCount - int64(keepCount)
	
	result := DB.Exec(`
		DELETE FROM events 
		WHERE id IN (
			SELECT id FROM events 
			ORDER BY timestamp ASC 
			LIMIT ?
		)
	`, deleteCount)

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Cleaned up %d old events, kept latest %d events", result.RowsAffected, keepCount)
	return nil
}