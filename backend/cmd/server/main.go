package main

import (
	"log"

	"miningpet/internal/handlers"
	"miningpet/internal/services"
	"miningpet/pkg/websocket"
	
	"github.com/gin-gonic/gin"
)

func main() {
	petService := services.NewPetService()
	petHandler := handlers.NewPetHandler(petService)
	hub := websocket.NewHub(petService)

	go hub.Run()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		api.POST("/pets", petHandler.CreatePet)
		api.GET("/pets", petHandler.GetAllPets)
		api.GET("/pets/:id", petHandler.GetPet)
		api.POST("/pets/:id/explore", petHandler.StartExploration)
		api.GET("/events", petHandler.GetEvents)
	}

	r.GET("/ws", hub.HandleWebSocket)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}