package models

import (
	"gorm.io/gorm"
	"time"
)

type Event struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Title       string          `json:"title" gorm:"not null"`
	Description string          `json:"description"`
	Location    string          `json:"location" gorm:"not null"`
	StartTime   time.Time       `json:"start_time" gorm:"not null"`
	OrganizerID uint            `json:"organizer_id" gorm:"not null"`
	Organizer   User            `json:"organizer" gorm:"foreignKey:OrganizerID"`
	Attendees   []EventAttendee `json:"attendees,omitempty" gorm:"foreignKey:EventID"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
}

type EventAttendee struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	EventID   uint           `json:"event_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Status    string         `json:"status" gorm:"default:'pending'"`           // pending, going, maybe, not_going
	Event     Event          `json:"event,omitempty" gorm:"foreignKey:EventID"` // ADD THIS LINE
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
