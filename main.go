package main

import (
    "eventplanner-backend/config"
    "eventplanner-backend/routes"
)

func main() {
    config.ConnectDatabase()
    router := routes.SetupRouter()
    router.Run(":8080")
}