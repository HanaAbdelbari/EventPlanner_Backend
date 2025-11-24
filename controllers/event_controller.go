package controllers

import (
    "EventPlanner_Backend/config"
    "EventPlanner_Backend/models"
    "net/http"

    "github.com/gin-gonic/gin"
)

//input from frontend
type CreateEventInput struct {
    Title       string `json:"title" binding:"required"`
    Date        string `json:"date" binding:"required"`
    Time        string `json:"time" binding:"required"`
    Location    string `json:"location" binding:"required"`
    Description string `json:"description"`
}


func CreateEvent(c *gin.Context) {
    userID := c.GetUint("user_id")

    var input CreateEventInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    event := models.Event{
        Title:       input.Title,
        Date:        input.Date,
        Time:        input.Time,
        Location:    input.Location,
        Description: input.Description,
        OrganizerID: userID,
    }

    if err := config.DB.Create(&event).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
        return
    }

    config.DB.Preload("Organizer").First(&event, event.ID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Event created successfully",
        "event":   event,
    })
}