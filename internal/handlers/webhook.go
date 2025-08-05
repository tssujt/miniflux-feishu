package handlers

import (
	"log"
	"net/http"

	"miniflux-feishu/internal/models"

	"github.com/gin-gonic/gin"
)

// FeishuServiceInterface defines the interface for Feishu service
type FeishuServiceInterface interface {
	SendEntryToFeishu(entry *models.WebhookEntry, feed *models.WebhookFeed, webhookURL string) error
}

type WebhookHandler struct {
	feishuService FeishuServiceInterface
}

func NewWebhookHandler(feishuService FeishuServiceInterface) *WebhookHandler {
	return &WebhookHandler{
		feishuService: feishuService,
	}
}

func (h *WebhookHandler) HandleMinifluxWebhook(c *gin.Context) {
	eventType := c.GetHeader("X-Miniflux-Event-Type")
	if eventType != "new_entries" {
		log.Printf("Ignoring event type: %s", eventType)
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// 获取飞书 webhook URL 参数
	webhookURL := c.Query("webhook_url")
	if webhookURL == "" {
		log.Printf("Missing webhook_url parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook_url parameter is required"})
		return
	}

	var webhookEvent models.WebhookNewEntriesEvent
	if err := c.ShouldBindJSON(&webhookEvent); err != nil {
		log.Printf("Failed to parse webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	log.Printf("Received %d new entries from feed: %s", len(webhookEvent.Entries), webhookEvent.Feed.Title)
	log.Printf("Using webhook URL: %s", webhookURL)

	for _, entry := range webhookEvent.Entries {
		if err := h.feishuService.SendEntryToFeishu(entry, webhookEvent.Feed, webhookURL); err != nil {
			log.Printf("Failed to send entry %d to Feishu: %v", entry.ID, err)
		} else {
			log.Printf("Successfully sent entry %d to Feishu", entry.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}
