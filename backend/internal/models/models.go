package models

import (
	"time"

	"gorm.io/gorm"
)

type Script struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	PlayerCount int            `gorm:"not null" json:"player_count"`
	Type        string         `gorm:"not null" json:"type"`
	Difficulty  string         `json:"difficulty"`
	Duration    int            `json:"duration"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Room struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Capacity  int            `gorm:"not null" json:"capacity"`
	Status    string         `gorm:"default:available" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Carpool struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ScriptID       uint           `gorm:"not null" json:"script_id"`
	Script         Script         `gorm:"foreignKey:ScriptID" json:"script"`
	RoomID         *uint          `json:"room_id"`
	Room           *Room          `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	HostName       string         `gorm:"not null" json:"host_name"`
	HostContact    string         `json:"host_contact"`
	CurrentPlayers int            `gorm:"not null;default:1" json:"current_players"`
	RequiredPlayers int           `gorm:"not null" json:"required_players"`
	Status         string         `gorm:"default:recruiting" json:"status"`
	StartTime      *time.Time     `json:"start_time"`
	Players        []Player       `gorm:"foreignKey:CarpoolID" json:"players"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Player struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CarpoolID   uint           `gorm:"not null" json:"carpool_id"`
	Name        string         `gorm:"not null" json:"name"`
	Contact     string         `json:"contact"`
	IsHost      bool           `gorm:"default:false" json:"is_host"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
