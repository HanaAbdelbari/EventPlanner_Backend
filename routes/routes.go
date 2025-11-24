package routes

import (
    "EventPlanner_Backend/controllers"
    "EventPlanner_Backend/middleware"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "time"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    api := r.Group("/api")
    //public paths don't need jwt and login
    api.POST("/signup", controllers.Signup)
    api.POST("/login", controllers.Login)

    //protected paths need login
    protected := api.Group("/")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/events", controllers.CreateEvent)
        protected.DELETE("/events/:id", controllers.DeleteEvent)
        //loading..........
    }

    return r
}