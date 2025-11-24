
package controllers

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/models"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
)

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

	location, _ := time.LoadLocation("UTC")
	dateTimeStr := input.Date + " " + input.Time // 2025-12-31 21:00
	startTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "صيغة التاريخ أو الوقت غير صحيحة"})
		return
	}

	event := models.Event{
		Title:       input.Title,
		Description: input.Description,
		Location:    input.Location,
		StartTime:   startTime,
		OrganizerID: userID,
	}

	result := config.DB.Create(&event)
	if result.Error != nil {
		log.Println("DATABASE ERROR:", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء الإيفنت", "details": result.Error.Error()})
		return
	}

	var fullEvent models.Event
	config.DB.Model(&models.Event{}).Where("id = ?", event.ID).Preload("Organizer").First(&fullEvent)

	c.JSON(http.StatusCreated, gin.H{
		"message": "تم إنشاء الإيفنت بنجاح",
		"event":   fullEvent,
	})
}

func DeleteEvent(c *gin.Context) {
	userID := c.GetUint("user_id")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صالح"})
		return
	}

	var event models.Event
	if err := config.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الإيفنت غير موجود"})
		return
	}

	if event.OrganizerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "لا يمكنك حذف إيفنت ليس ملكك"})
		return
	}

	config.DB.Delete(&event)
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الإيفنت بنجاح"})
}

func GetMyEvents(c *gin.Context) {
	userID := c.GetUint("user_id")

	var events []models.Event
	config.DB.Preload("Organizer").Where("organizer_id = ?", userID).Find(&events)

	c.JSON(http.StatusOK, events)
}

func GetInvitedEvents(c *gin.Context) {


	var events []models.Event
	config.DB.Preload("Organizer").Find(&events)

	c.JSON(http.StatusOK, events)
}

func GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صحيح"})
		return
	}

	var event models.Event
	result := config.DB.Preload("Organizer").First(&event, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الإيفنت غير موجود"})
		return
	}

	c.JSON(http.StatusOK, event)
}