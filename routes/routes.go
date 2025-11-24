package routes

import (
    "EventPlanner_Backend/controllers"
    "EventPlanner_Backend/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")

    // Public routes (no auth)
    api.POST("/signup", controllers.Signup)
    api.POST("/login", controllers.Login)

    // Protected routes (with auth)
    protected := api.Group("/")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/events", controllers.CreateEvent)
        protected.DELETE("/events/:id", controllers.DeleteEvent)
       protected.GET("/events/organized", controllers.GetMyEvents)
       protected.GET("/events/invited", controllers.GetInvitedEvents)
       protected.GET("/events/:id", controllers.GetEventByID)
    }

    return r
}