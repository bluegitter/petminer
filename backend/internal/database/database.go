package database

import (
	"database/sql"
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

	// 连接SQLite数据库 - 添加性能优化参数
	dbPath += "?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=memory"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
		// 优化GORM配置
		SkipDefaultTransaction: true, // 手动控制事务
		PrepareStmt:           true,  // 预编译语句
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db

	// 配置连接池 - 优化并发性能
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// SQLite连接池配置 - 提高并发能力
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 执行性能优化PRAGMA语句
	if err := optimizeSQLiteSettings(sqlDB); err != nil {
		return fmt.Errorf("failed to optimize SQLite settings: %w", err)
	}

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
	
	// 初始化批量写入管理器
	InitializeBatchManagers()
	log.Println("Batch write managers initialized")
	
	return nil
}

// Close 关闭数据库连接
func Close() error {
	// 先关闭批量写入管理器
	CloseBatchManagers()
	
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

// optimizeSQLiteSettings 优化SQLite性能设置
func optimizeSQLiteSettings(sqlDB *sql.DB) error {
	// 执行性能优化PRAGMA语句
	pragmas := []string{
		"PRAGMA busy_timeout = 10000",        // 设置忙等待超时
		"PRAGMA foreign_keys = ON",           // 启用外键约束
		"PRAGMA wal_autocheckpoint = 1000",   // WAL自动检查点
		"PRAGMA mmap_size = 268435456",       // 256MB内存映射
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			log.Printf("Warning: failed to execute pragma '%s': %v", pragma, err)
		}
	}

	log.Println("SQLite performance settings optimized")
	return nil
}

// BatchWrite 批量写入操作接口
type BatchWrite interface {
	Execute(tx *gorm.DB) error
}

// BatchWriteManager 批量写入管理器
type BatchWriteManager struct {
	writeQueue chan BatchWrite
	batchSize  int
	flushTime  time.Duration
	quit       chan bool
}

// NewBatchWriteManager 创建批量写入管理器
func NewBatchWriteManager(batchSize int, flushTime time.Duration) *BatchWriteManager {
	manager := &BatchWriteManager{
		writeQueue: make(chan BatchWrite, batchSize*2),
		batchSize:  batchSize,
		flushTime:  flushTime,
		quit:       make(chan bool),
	}
	
	go manager.processBatchWrites()
	return manager
}

// AddWrite 添加写入操作到队列
func (bm *BatchWriteManager) AddWrite(write BatchWrite) {
	select {
	case bm.writeQueue <- write:
	default:
		log.Println("Warning: write queue is full, dropping write operation")
	}
}

// Stop 停止批量写入管理器
func (bm *BatchWriteManager) Stop() {
	close(bm.quit)
}

// processBatchWrites 处理批量写入
func (bm *BatchWriteManager) processBatchWrites() {
	ticker := time.NewTicker(bm.flushTime)
	defer ticker.Stop()

	var batch []BatchWrite

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := DB.Transaction(func(tx *gorm.DB) error {
			for _, write := range batch {
				if err := write.Execute(tx); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			log.Printf("Error executing batch write: %v", err)
		}

		batch = batch[:0] // 清空批次但保留容量
	}

	for {
		select {
		case write := <-bm.writeQueue:
			batch = append(batch, write)
			if len(batch) >= bm.batchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-bm.quit:
			flush() // 最后一次刷新
			return
		}
	}
}