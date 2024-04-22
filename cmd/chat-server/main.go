package main

import (
	"chat/internal/api/handlers"
	"chat/internal/api/middleware"
	"chat/internal/api/routes"
	"chat/internal/data/db"
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	cleanup := middleware.InitTracerAuto()
	defer cleanup(context.Background())
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	if err := db.InitDB(); err != nil {
		panic(err)
	}
	defer db.CloseDB()

	router := routes.SetupRoutes()
	go router.Run(":8181")

	// Start the server
	http.HandleFunc("/ws", handlers.HandleWebsocketConnection)
	log.Println("WS server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
