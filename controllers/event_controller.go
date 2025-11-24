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

func DeleteEvent(c *gin.Context) {

    userID := c.GetUint("user_id")

    eventID := c.Param("id")

    var event models.Event

    if err := config.DB.First(&event, eventID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
        return
    }

    if event.OrganizerID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete events you created"})
        return
    }

    if err := config.DB.Delete(&event).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Event deleted successfully",
    })
}