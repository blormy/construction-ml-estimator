package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// реализация valuer и scanner для работы gorm с типом jsonb
type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = make([]byte, len(v))
		copy(*j, v)
	case string:
		*j = JSONB(v)
	default:
		return fmt.Errorf("JSONB: неподдерживаемый тип %T", value)
	}
	return nil
}

type User struct {
	ID           uint      `gorm:"primarykey;column:id"`
	Email        string    `gorm:"unique;not null;column:email"`
	PasswordHash string    `gorm:"not null;column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

type Project struct {
	ID             uint            `gorm:"primarykey;column:id"`
	UserID         uint            `gorm:"not null;index;column:user_id"`
	TeamID         *uint           `gorm:"index;column:team_id"`
	Name           string          `gorm:"not null;column:name"`
	Description    string          `gorm:"column:description"`
	EquipmentItems []EquipmentItem `gorm:"foreignKey:ProjectID"`
	CreatedAt      time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

type EquipmentItem struct {
	ID            uint      `gorm:"primarykey;column:id"`
	ProjectID     uint      `gorm:"not null;index;column:project_id"`
	EquipmentType string    `gorm:"not null;column:equipment_type"`
	Parameters    JSONB     `gorm:"column:parameters;type:jsonb"`
	WeightResult  *float64  `gorm:"column:weight_result"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
}

// модель команды для совместной работы над проектами
type Team struct {
	ID        uint         `gorm:"primarykey;column:id"`
	Name      string       `gorm:"not null;column:name"`
	OwnerID   uint         `gorm:"not null;index;column:owner_id"`
	Members   []TeamMember `gorm:"foreignKey:TeamID"`
	CreatedAt time.Time    `gorm:"column:created_at;autoCreateTime"`
}

// участник команды
type TeamMember struct {
	ID     uint   `gorm:"primarykey;column:id"`
	TeamID uint   `gorm:"not null;uniqueIndex:idx_team_user;column:team_id"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_team_user;column:user_id"`
	Role   string `gorm:"not null;default:'member';column:role"`
}
