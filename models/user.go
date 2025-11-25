package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`

	Events []Event `json:"events" gorm:"foreignKey:OrganizerID"`
	AttendedEvents []EventAttendee `json:"attended_events" gorm:"foreignKey:UserID"`
}