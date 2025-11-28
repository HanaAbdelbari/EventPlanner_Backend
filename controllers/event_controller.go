package controllers

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateEventInput struct {
	Title       string   `json:"title" binding:"required"`
	Date        string   `json:"date" binding:"required"`
	Time        string   `json:"time" binding:"required"`
	Location    string   `json:"location" binding:"required"`
	Description string   `json:"description"`
	Invitees    []string `json:"invitees"`
}

func CreateEvent(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, _ := time.LoadLocation("UTC")
	dateTimeStr := input.Date + " " + input.Time
	startTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date or time format"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	// Invite users if provided
	if len(input.Invitees) > 0 {
		for _, email := range input.Invitees {
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}

			var invitedUser models.User
			if err := config.DB.Where("email = ?", email).First(&invitedUser).Error; err != nil {
				log.Println("User not found:", email)
				continue
			}

			if invitedUser.ID == userID {
				continue
			}

			var existingInvite models.EventAttendee
			if config.DB.Where("event_id = ? AND user_id = ?", event.ID, invitedUser.ID).First(&existingInvite).Error == nil {
				continue
			}

			invitation := models.EventAttendee{
				EventID: event.ID,
				UserID:  invitedUser.ID,
				Status:  "pending",
			}
			config.DB.Create(&invitation)
		}
	}

	var fullEvent models.Event
	config.DB.Model(&models.Event{}).Where("id = ?", event.ID).Preload("Organizer").First(&fullEvent)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
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

	var formattedEvents []map[string]interface{}
	for _, event := range events {
		formattedEvents = append(formattedEvents, map[string]interface{}{
			"id":           event.ID,
			"title":        event.Title,
			"description":  event.Description,
			"date":         event.StartTime.Format("2006-01-02"),
			"time":         event.StartTime.Format("15:04"),
			"location":     event.Location,
			"organizer_id": event.OrganizerID,
			"created_at":   event.CreatedAt,
			"role":         "organizer",
			"status":       "going",
		})
	}

	c.JSON(http.StatusOK, formattedEvents)
}

func GetInvitedEvents(c *gin.Context) {
	userID := c.GetUint("user_id")

	var attendees []models.EventAttendee
	config.DB.Where("user_id = ?", userID).Preload("Event").Preload("Event.Organizer").Find(&attendees)

	var formattedEvents []map[string]interface{}
	for _, attendee := range attendees {
		formattedEvents = append(formattedEvents, map[string]interface{}{
			"id":           attendee.Event.ID,
			"title":        attendee.Event.Title,
			"description":  attendee.Event.Description,
			"date":         attendee.Event.StartTime.Format("2006-01-02"),
			"time":         attendee.Event.StartTime.Format("15:04"),
			"location":     attendee.Event.Location,
			"organizer_id": attendee.Event.OrganizerID,
			"created_at":   attendee.Event.CreatedAt,
			"role":         "attendee",
			"status":       attendee.Status,
		})
	}

	c.JSON(http.StatusOK, formattedEvents)
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

func GetAllMyEvents(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Get organized events
	var organizedEvents []models.Event
	config.DB.Preload("Organizer").Where("organizer_id = ?", userID).Find(&organizedEvents)

	var formattedEvents []map[string]interface{}

	// Add organized events
	for _, event := range organizedEvents {
		formattedEvents = append(formattedEvents, map[string]interface{}{
			"id":           event.ID,
			"title":        event.Title,
			"description":  event.Description,
			"date":         event.StartTime.Format("2006-01-02"),
			"time":         event.StartTime.Format("15:04"),
			"location":     event.Location,
			"organizer_id": event.OrganizerID,
			"created_at":   event.CreatedAt,
			"role":         "organizer",
			"status":       "going",
		})
	}

	// Get invited events
	var attendees []models.EventAttendee
	config.DB.Where("user_id = ?", userID).Preload("Event").Preload("Event.Organizer").Find(&attendees)

	for _, attendee := range attendees {
		formattedEvents = append(formattedEvents, map[string]interface{}{
			"id":           attendee.Event.ID,
			"title":        attendee.Event.Title,
			"description":  attendee.Event.Description,
			"date":         attendee.Event.StartTime.Format("2006-01-02"),
			"time":         attendee.Event.StartTime.Format("15:04"),
			"location":     attendee.Event.Location,
			"organizer_id": attendee.Event.OrganizerID,
			"created_at":   attendee.Event.CreatedAt,
			"role":         "attendee",
			"status":       attendee.Status,
		})
	}

	c.JSON(http.StatusOK, formattedEvents)
}


func InviteUser(c *gin.Context) {
	organizerID := c.GetUint("user_id")

	// Get event_id and email from request body (NOT from URL)
	var input struct {
		EventID uint   `json:"event_id" binding:"required"`
		Email   string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if event exists and user is the organizer
	var event models.Event
	if err := config.DB.First(&event, input.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.OrganizerID != organizerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only invite users to your own events"})
		return
	}

	var invitedUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&invitedUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found with this email"})
		return
	}

	// Check if user is the organizer
	if invitedUser.ID == organizerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot invite yourself"})
		return
	}

	// Check if already invited
	var existingInvite models.EventAttendee
	result := config.DB.Where("event_id = ? AND user_id = ?", input.EventID, invitedUser.ID).First(&existingInvite)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already invited to this event"})
		return
	}

	// Create invitation
	invitation := models.EventAttendee{
		EventID: input.EventID,
		UserID:  invitedUser.ID,
		Status:  "pending",
	}

	if err := config.DB.Create(&invitation).Error; err != nil {
		log.Println("DATABASE ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send invitation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invitation sent successfully",
	})
}

// Respond to invitation (Accept/Decline)
func RespondToInvitation(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Get event_id and status from request body (NOT from URL)
	var input struct {
		EventID uint   `json:"event_id" binding:"required"`
		Status  string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"going":     true,
		"maybe":     true,
		"not_going": true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be: going, maybe, or not_going"})
		return
	}

	// Find invitation
	var invitation models.EventAttendee
	if err := config.DB.Where("event_id = ? AND user_id = ?", input.EventID, userID).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found for this event"})
		return
	}

	// Update status
	invitation.Status = input.Status
	if err := config.DB.Save(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

func GetEventAttendees(c *gin.Context) {
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var attendees []models.EventAttendee
	config.DB.Where("event_id = ?", eventID).Preload("User").Find(&attendees)

	// Format for frontend - FIXED
	var formattedAttendees []map[string]interface{}
	for _, attendee := range attendees {
		role := "attendee"
		if attendee.UserID == event.OrganizerID {
			role = "organizer"
		}

		formattedAttendees = append(formattedAttendees, map[string]interface{}{
			"id":     attendee.User.ID,
			"name":   attendee.User.Name,
			"email":  attendee.User.Email,
			"role":   role,
			"status": attendee.Status,
		})
	}

	c.JSON(http.StatusOK, formattedAttendees)
}


func SearchEvents(c *gin.Context) {
    userID := c.GetUint("user_id")

    keyword := strings.TrimSpace(c.Query("keyword"))
    startDateStr := c.Query("start_date")
    endDateStr := c.Query("end_date")
    roleParam := strings.ToLower(strings.TrimSpace(c.Query("role")))
    status := strings.ToLower(strings.TrimSpace(c.Query("status")))

    var results []map[string]interface{}
    seen := make(map[uint]bool)

    // توقيت مصر
    loc, _ := time.LoadLocation("Africa/Cairo")

    var startTime, endTime time.Time

    if startDateStr != "" {
        t, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date"})
            return
        }
        startTime = t.In(time.UTC)
    }

    if endDateStr != "" {
        t, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date"})
            return
        }
        // نهاية اليوم بتوقيت مصر (23:59:59) → نحوله لـ UTC
        endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, loc)
        endTime = endOfDay.In(time.UTC)
    }

    includeOrg := roleParam == "" || roleParam == "organizer" || roleParam == "all"
    includeAtt := roleParam == "" || roleParam == "attendee" || roleParam == "all"
    if status != "" && status != "going" {
        includeOrg = false
    }

    // Organized Events
    if includeOrg {
        q := config.DB.Preload("Organizer").Where("organizer_id = ?", userID)

        if keyword != "" {
            q = q.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
        }
        if !startTime.IsZero() {
            q = q.Where("start_time >= ?", startTime)
        }
        if !endTime.IsZero() {
            q = q.Where("start_time <= ?", endTime)
        }

        var events []models.Event
        q.Find(&events)
        for _, e := range events {
            if seen[e.ID] { continue }
            seen[e.ID] = true
            results = append(results, map[string]interface{}{
                "id":           e.ID,
                "title":        e.Title,
                "description":  e.Description,
                "date":         e.StartTime.In(loc).Format("2006-01-02"),
                "time":         e.StartTime.In(loc).Format("15:04"),
                "location":     e.Location,
                "organizer_id": e.OrganizerID,
                "created_at":   e.CreatedAt,
                "role":         "organizer",
                "status":       "going",
            })
        }
    }

    // Attended Events
    if includeAtt {
        q := config.DB.Preload("Event").Preload("Event.Organizer").
            Joins("JOIN events ON events.id = event_attendees.event_id").
            Where("event_attendees.user_id = ?", userID)

        if keyword != "" {
            q = q.Where("events.title LIKE ? OR events.description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
        }
        if !startTime.IsZero() {
            q = q.Where("events.start_time >= ?", startTime)
        }
        if !endTime.IsZero() {
            q = q.Where("events.start_time <= ?", endTime)
        }
        if status != "" {
            q = q.Where("event_attendees.status = ?", status)
        }

        var attendees []models.EventAttendee
        q.Find(&attendees)
        for _, a := range attendees {
            if seen[a.EventID] { continue }
            seen[a.EventID] = true
            eventTime := a.Event.StartTime.In(loc)
            results = append(results, map[string]interface{}{
                "id":           a.Event.ID,
                "title":        a.Event.Title,
                "description":  a.Event.Description,
                "date":         eventTime.Format("2006-01-02"),
                "time":         eventTime.Format("15:04"),
                "location":     a.Event.Location,
                "organizer_id": a.Event.OrganizerID,
                "created_at":   a.Event.CreatedAt,
                "role":         "attendee",
                "status":       a.Status,
            })
        }
    }

    c.JSON(http.StatusOK, results)
}