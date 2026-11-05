package db

import (
	"constructmaterialai/internal/models"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// глобальный экземпляр соединения с бд
var DB *gorm.DB

// инициализация подключения к базе данных
func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL не задана")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// создаем схему sync, если она не существует
	if err := DB.Exec("CREATE SCHEMA IF NOT EXISTS sync").Error; err != nil {
		return fmt.Errorf("ошибка создания схемы sync: %w", err)
	}

	// устанавливаем путь поиска (search_path) на sync
	if err := DB.Exec("SET search_path TO sync,public").Error; err != nil {
		return fmt.Errorf("ошибка установки search_path: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("ошибка получения sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	return nil
}

// автомиграция всех таблиц
func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.EquipmentItem{},
		&models.Team{},
		&models.TeamMember{},
	)
}
