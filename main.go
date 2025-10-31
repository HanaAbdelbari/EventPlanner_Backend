package main

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/routes"
)

func main() {
	config.ConnectDB()
	defer func() {
		sqlDB, _ := config.DB.DB()
		sqlDB.Close()
	}()

	routes.SetupRouter().Run(":8080")
}
