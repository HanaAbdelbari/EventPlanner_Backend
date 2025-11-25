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

	Organizer   User              `json:"organizer" gorm:"foreignKey:OrganizerID"`
	Attendees   []EventAttendee   `json:"attendees" gorm:"foreignKey:EventID"`
}

type EventAttendee struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	EventID   uint      `gorm:"index"`
	UserID    uint      `gorm:"index"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	CreatedAt time.Time
}

func (Event) TableName() string {
	return "events"
}