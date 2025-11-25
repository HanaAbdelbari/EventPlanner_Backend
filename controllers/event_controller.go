package controllers

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
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
	userID := c.GetUint("user_id")

	// Get event IDs where user is invited
	var attendees []models.EventAttendee
	config.DB.Where("user_id = ?", userID).Find(&attendees)

	// Extract event IDs
	var eventIDs []uint
	for _, attendee := range attendees {
		eventIDs = append(eventIDs, attendee.EventID)
	}

	// Get events
	var events []models.Event
	if len(eventIDs) > 0 {
		config.DB.Preload("Organizer").Where("id IN ?", eventIDs).Find(&events)
	}

	c.JSON(http.StatusOK, events)
}

func GetEventByID(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صحيح"})
		return
	}

	var event models.Event
	result := config.DB.Preload("Organizer").Preload("Attendees.User").First(&event, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الإيفنت غير موجود"})
		return
	}

	var userRole string
	if event.OrganizerID == currentUserID {
		userRole = "organizer"
	} else {

		for _, attendee := range event.Attendees {
			if attendee.UserID == currentUserID {
				userRole = "attendee"
				break
			}
		}
		if userRole == "" {
			userRole = "visitor"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"event":     event,
		"user_role": userRole,
	})

}

// Invite a user to an event
func InviteUser(c *gin.Context) {
	organizerID := c.GetUint("user_id")
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صالح"})
		return
	}

	// Check if event exists and user is the organizer
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الإيفنت غير موجود"})
		return
	}

	if event.OrganizerID != organizerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "يمكنك فقط دعوة مستخدمين لإيفنتاتك"})
		return
	}

	// Get user_id or email from request
	var input struct {
		UserID uint   `json:"user_id"`
		Email  string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var invitedUser models.User

	// Find user by ID or email
	if input.UserID != 0 {
		if err := config.DB.First(&invitedUser, input.UserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
			return
		}
	} else if input.Email != "" {
		if err := config.DB.Where("email = ?", input.Email).First(&invitedUser).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود بهذا البريد الإلكتروني"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يجب توفير user_id أو email"})
		return
	}

	// Check if user is the organizer
	if invitedUser.ID == organizerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لا يمكنك دعوة نفسك"})
		return
	}

	// Check if already invited
	var existingInvite models.EventAttendee
	result := config.DB.Where("event_id = ? AND user_id = ?", eventID, invitedUser.ID).First(&existingInvite)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "المستخدم مدعو بالفعل لهذا الإيفنت"})
		return
	}

	// Create invitation
	invitation := models.EventAttendee{
		EventID: uint(eventID),
		UserID:  invitedUser.ID,
		Status:  "pending",
	}

	if err := config.DB.Create(&invitation).Error; err != nil {
		log.Println("DATABASE ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إرسال الدعوة"})
		return
	}

	// Load user info
	config.DB.Preload("User").First(&invitation, invitation.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":    "تم إرسال الدعوة بنجاح",
		"invitation": invitation,
	})
}

// Respond to invitation (Accept/Decline)
func RespondToInvitation(c *gin.Context) {
	userID := c.GetUint("user_id")
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صالح"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	if input.Status != "accepted" && input.Status != "declined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الحالة يجب أن تكون accepted أو declined"})
		return
	}

	// Find invitation
	var invitation models.EventAttendee
	if err := config.DB.Where("event_id = ? AND user_id = ?", eventID, userID).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "لم يتم العثور على دعوة لهذا الإيفنت"})
		return
	}

	// Update status
	invitation.Status = input.Status
	if err := config.DB.Save(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تحديث الحالة"})
		return
	}

	message := "تم قبول الدعوة بنجاح"
	if input.Status == "declined" {
		message = "تم رفض الدعوة"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"invitation": invitation,
	})
}

// Get all attendees for an event
func GetEventAttendees(c *gin.Context) {
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الإيفنت غير صالح"})
		return
	}

	// Check if event exists
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الإيفنت غير موجود"})
		return
	}

	// Get all attendees
	var attendees []models.EventAttendee
	config.DB.Where("event_id = ?", eventID).Preload("User").Find(&attendees)

	c.JSON(http.StatusOK, gin.H{
		"event_id":  eventID,
		"attendees": attendees,
		"total":     len(attendees),
	})
}
