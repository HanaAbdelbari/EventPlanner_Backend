
package main

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/models"
	"EventPlanner_Backend/routes"
	"log"
)

func main() {
	config.ConnectDB()

	// Auto migrate tables
	if err := config.DB.AutoMigrate(&models.User{}, &models.Event{}, &models.EventParticipant{}); err != nil {
		log.Fatalln("AutoMigrate failed:", err)
	}

	r := routes.SetupRouter()
	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}
