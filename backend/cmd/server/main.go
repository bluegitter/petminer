package main

import (
	"log"
	"os"

	"miningpet/internal/handlers"
	"miningpet/internal/services"
	"miningpet/pkg/websocket"
	
	"github.com/gin-gonic/gin"
)

// 构建时注入的版本信息
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 设置生产环境模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

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

	// 获取端口配置，默认8081用于容器内部
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// 添加版本信息端点
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
		})
	})

	log.Printf("PetMiner Server v%s", Version)
	log.Printf("Build: %s (%s)", BuildTime, GitCommit)
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}