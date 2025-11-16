package main

import (
	"fmt"
	"log"

	"HOSEROF_API/config"
	"HOSEROF_API/routes"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("warning: .env not found or failed to load")
	}

	config.InitFirebase()
	defer config.DB.Close()

	fmt.Println("Server running on :3000")
	router := routes.SetupRouter()
	if err := router.Run(":3000"); err != nil {
		log.Fatal("server failed:", err)
	}
}
