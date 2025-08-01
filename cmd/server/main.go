package main

import (
	"log"
	"net/http"
	"os"

	"miniflux-feishu/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting Miniflux-Feishu Integration Service...")

	r, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// 从环境变量获取端口，默认8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Webhook endpoint: http://localhost:%s/webhook/miniflux?webhook_url=YOUR_FEISHU_WEBHOOK_URL", port)
	log.Printf("Health check endpoint: http://localhost:%s/health", port)

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(webhookHandler *handlers.WebhookHandler) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "miniflux-feishu",
		})
	})

	webhook := r.Group("/webhook")
	webhook.Use(gin.Logger(), gin.Recovery())

	webhook.POST("/miniflux", webhookHandler.HandleMinifluxWebhook)

	return r
}
