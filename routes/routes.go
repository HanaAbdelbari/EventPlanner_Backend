package routes

import (
	"EventPlanner_Backend/controllers"
	"EventPlanner_Backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS Configuration - More specific
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Server is running!",
		})
	})

	api := r.Group("/api")

	// Public routes (no auth)
	api.POST("/signup", controllers.Signup)
	api.POST("/login", controllers.Login)

	// Protected routes (with auth)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Events - FIXED ROUTES
		protected.POST("/events", controllers.CreateEvent)
		protected.DELETE("/events/:id", controllers.DeleteEvent)
		protected.GET("/events/my", controllers.GetMyEvents) // Changed from /organized
		protected.GET("/events/invited", controllers.GetInvitedEvents)
		protected.GET("/events/all", controllers.GetAllMyEvents) // ADDED - for dashboard
		protected.GET("/events/search", controllers.SearchEvents)
		protected.GET("/events/:id", controllers.GetEventByID)
		protected.GET("/events/:id/attendees", controllers.GetEventAttendees)

		// Invitations - FIXED ROUTES
		protected.POST("/events/invite", controllers.InviteUser)           // Changed from /:id/invite
		protected.PUT("/events/response", controllers.RespondToInvitation) // Changed from /:id/rsvp
	}

	return r
}
