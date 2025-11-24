package models

import (
	"gorm.io/gorm"
	"time"
)

type Event struct {
	gorm.Model
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	Location    string    `json:"location" gorm:"not null"`
	StartTime   time.Time `json:"start_time" gorm:"not null"`
	OrganizerID uint      `json:"organizer_id" gorm:"not null"`
	Organizer User `json:"organizer" gorm:"foreignKey:OrganizerID;references:ID"`
}