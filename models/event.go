// models/event.go
package models

import "gorm.io/gorm"

type Event struct {
    gorm.Model
    Title       string `json:"title" gorm:"not null"`
    Date        string `json:"date" gorm:"not null"`
    Time        string `json:"time" gorm:"not null"`
    Location    string `json:"location" gorm:"not null"`
    Description string `json:"description" gorm:"type:text"`

    OrganizerID uint `json:"organizer_id" gorm:"not null"`
    Organizer   User `json:"organizer" gorm:"foreignKey:OrganizerID"`

}