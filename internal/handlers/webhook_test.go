package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux-feishu/internal/models"
	"miniflux-feishu/internal/services"

	"github.com/gin-gonic/gin"
)

// MockFeishuService is a mock implementation of FeishuServiceInterface for testing
type MockFeishuService struct {
	sendEntryToFeishuFunc func(entry *models.WebhookEntry, feed *models.WebhookFeed, webhookURL string) error
	callCount             int
	lastEntry             *models.WebhookEntry
	lastFeed              *models.WebhookFeed
	lastWebhookURL        string
}

func (m *MockFeishuService) SendEntryToFeishu(entry *models.WebhookEntry, feed *models.WebhookFeed, webhookURL string) error {
	m.callCount++
	m.lastEntry = entry
	m.lastFeed = feed
	m.lastWebhookURL = webhookURL

	if m.sendEntryToFeishuFunc != nil {
		return m.sendEntryToFeishuFunc(entry, feed, webhookURL)
	}
	return nil
}

func TestWebhookHandler_HandleMinifluxWebhook_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	// Create test payload
	payload := `{
		"event_type": "new_entries",
		"feed": {
			"id": 8,
			"user_id": 1,
			"feed_url": "https://example.org/feed.xml",
			"site_url": "https://example.org",
			"title": "Example website",
			"checked_at": "2023-09-10T12:48:43.428196-07:00"
		},
		"entries": [
			{
				"id": 231,
				"user_id": 1,
				"feed_id": 3,
				"status": "unread",
				"hash": "1163a93ef12741b558a3b86d7e975c4c1de0152f3439915ed185eb460e5718d7",
				"title": "Example",
				"url": "https://example.org/article",
				"comments_url": "",
				"published_at": "2023-08-17T19:29:22Z",
				"created_at": "2023-09-10T12:48:43.428196-07:00",
				"changed_at": "2023-09-10T12:48:43.428196-07:00",
				"content": "<p>Some HTML content</p>",
				"share_code": "",
				"starred": false,
				"reading_time": 1,
				"enclosures": [{
					"id": 158,
					"user_id": 1,
					"entry_id": 231,
					"url": "https://example.org/podcast.mp3",
					"mime_type": "audio/mpeg",
					"size": 63451045,
					"media_progression": 0
				}],
				"tags": ["Some category", "Another label"]
			}
		]
	}`

	// Create HTTP request
	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Webhook processed successfully" {
		t.Errorf("Expected success message, got %v", response["message"])
	}

	// Verify mock service was called
	if mockService.callCount != 1 {
		t.Errorf("Expected FeishuService to be called 1 time, got %d", mockService.callCount)
	}

	// Verify correct parameters were passed to FeishuService
	if mockService.lastWebhookURL != "https://hooks.example.com/webhook" {
		t.Errorf("Expected webhook URL 'https://hooks.example.com/webhook', got '%s'", mockService.lastWebhookURL)
	}

	if mockService.lastEntry.ID != 231 {
		t.Errorf("Expected entry ID 231, got %d", mockService.lastEntry.ID)
	}

	if mockService.lastEntry.Title != "Example" {
		t.Errorf("Expected entry title 'Example', got '%s'", mockService.lastEntry.Title)
	}

	if mockService.lastFeed.ID != 8 {
		t.Errorf("Expected feed ID 8, got %d", mockService.lastFeed.ID)
	}

	if mockService.lastFeed.Title != "Example website" {
		t.Errorf("Expected feed title 'Example website', got '%s'", mockService.lastFeed.Title)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_WrongEventType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	payload := `{"event_type": "other_event"}`
	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "other_event")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	// Should return OK but ignore the event
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Event ignored" {
		t.Errorf("Expected 'Event ignored' message, got %v", response["message"])
	}

	// FeishuService should not be called
	if mockService.callCount != 0 {
		t.Errorf("Expected FeishuService not to be called, but it was called %d times", mockService.callCount)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_MissingWebhookURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	payload := `{"event_type": "new_entries"}`
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(payload)) // No webhook_url parameter
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "webhook_url parameter is required" {
		t.Errorf("Expected webhook_url error message, got %v", response["error"])
	}

	// FeishuService should not be called
	if mockService.callCount != 0 {
		t.Errorf("Expected FeishuService not to be called, but it was called %d times", mockService.callCount)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	payload := `{invalid json`
	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Invalid payload" {
		t.Errorf("Expected 'Invalid payload' error message, got %v", response["error"])
	}

	// FeishuService should not be called
	if mockService.callCount != 0 {
		t.Errorf("Expected FeishuService not to be called, but it was called %d times", mockService.callCount)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_FeishuServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service that returns an error
	mockService := &MockFeishuService{
		sendEntryToFeishuFunc: func(entry *models.WebhookEntry, feed *models.WebhookFeed, webhookURL string) error {
			return fmt.Errorf("feishu service error")
		},
	}
	handler := NewWebhookHandler(mockService)

	payload := `{
		"event_type": "new_entries",
		"feed": {
			"id": 8,
			"user_id": 1,
			"feed_url": "https://example.org/feed.xml",
			"site_url": "https://example.org",
			"title": "Example website",
			"checked_at": "2023-09-10T12:48:43.428196-07:00"
		},
		"entries": [
			{
				"id": 231,
				"user_id": 1,
				"feed_id": 3,
				"status": "unread",
				"hash": "1163a93ef12741b558a3b86d7e975c4c1de0152f3439915ed185eb460e5718d7",
				"title": "Example",
				"url": "https://example.org/article",
				"comments_url": "",
				"published_at": "2023-08-17T19:29:22Z",
				"created_at": "2023-09-10T12:48:43.428196-07:00",
				"changed_at": "2023-09-10T12:48:43.428196-07:00",
				"content": "<p>Some HTML content</p>",
				"share_code": "",
				"starred": false,
				"reading_time": 1,
				"enclosures": [],
				"tags": []
			}
		]
	}`

	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	// Should still return OK even if FeishuService fails (error is logged)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Webhook processed successfully" {
		t.Errorf("Expected success message, got %v", response["message"])
	}

	// FeishuService should have been called
	if mockService.callCount != 1 {
		t.Errorf("Expected FeishuService to be called 1 time, got %d", mockService.callCount)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_MultipleEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	payload := `{
		"event_type": "new_entries",
		"feed": {
			"id": 8,
			"user_id": 1,
			"feed_url": "https://example.org/feed.xml",
			"site_url": "https://example.org",
			"title": "Example website",
			"checked_at": "2023-09-10T12:48:43.428196-07:00"
		},
		"entries": [
			{
				"id": 231,
				"user_id": 1,
				"feed_id": 3,
				"status": "unread",
				"hash": "hash1",
				"title": "Example 1",
				"url": "https://example.org/article1",
				"comments_url": "",
				"published_at": "2023-08-17T19:29:22Z",
				"created_at": "2023-09-10T12:48:43.428196-07:00",
				"changed_at": "2023-09-10T12:48:43.428196-07:00",
				"content": "<p>Content 1</p>",
				"share_code": "",
				"starred": false,
				"reading_time": 1,
				"enclosures": [],
				"tags": []
			},
			{
				"id": 232,
				"user_id": 1,
				"feed_id": 3,
				"status": "unread",
				"hash": "hash2",
				"title": "Example 2",
				"url": "https://example.org/article2",
				"comments_url": "",
				"published_at": "2023-08-17T20:29:22Z",
				"created_at": "2023-09-10T12:48:43.428196-07:00",
				"changed_at": "2023-09-10T12:48:43.428196-07:00",
				"content": "<p>Content 2</p>",
				"share_code": "",
				"starred": false,
				"reading_time": 2,
				"enclosures": [],
				"tags": []
			}
		]
	}`

	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// FeishuService should be called once for each entry
	if mockService.callCount != 2 {
		t.Errorf("Expected FeishuService to be called 2 times, got %d", mockService.callCount)
	}
}

func TestWebhookHandler_HandleMinifluxWebhook_MissingEventTypeHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockFeishuService{}
	handler := NewWebhookHandler(mockService)

	payload := `{"event_type": "new_entries"}`
	req := httptest.NewRequest("POST", "/webhook?webhook_url=https://hooks.example.com/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-Miniflux-Event-Type header

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	// Should return OK but ignore the event (empty string != "new_entries")
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Event ignored" {
		t.Errorf("Expected 'Event ignored' message, got %v", response["message"])
	}

	// FeishuService should not be called
	if mockService.callCount != 0 {
		t.Errorf("Expected FeishuService not to be called, but it was called %d times", mockService.callCount)
	}
}

// Test that verifies the actual interface with real FeishuService (but without making real HTTP calls)
func TestWebhookHandler_IntegrationWithRealService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use real FeishuService
	realService := services.NewFeishuService()
	handler := NewWebhookHandler(realService)

	payload := `{
		"event_type": "new_entries",
		"feed": {
			"id": 8,
			"user_id": 1,
			"feed_url": "https://example.org/feed.xml",
			"site_url": "https://example.org",
			"title": "Example website",
			"checked_at": "2023-09-10T12:48:43.428196-07:00"
		},
		"entries": [
			{
				"id": 231,
				"user_id": 1,
				"feed_id": 3,
				"status": "unread",
				"hash": "1163a93ef12741b558a3b86d7e975c4c1de0152f3439915ed185eb460e5718d7",
				"title": "Example",
				"url": "https://example.org/article",
				"comments_url": "",
				"published_at": "2023-08-17T19:29:22Z",
				"created_at": "2023-09-10T12:48:43.428196-07:00",
				"changed_at": "2023-09-10T12:48:43.428196-07:00",
				"content": "<p>Some HTML content</p>",
				"share_code": "",
				"starred": false,
				"reading_time": 1,
				"enclosures": [{
					"id": 158,
					"user_id": 1,
					"entry_id": 231,
					"url": "https://example.org/podcast.mp3",
					"mime_type": "audio/mpeg",
					"size": 63451045,
					"media_progression": 0
				}],
				"tags": ["Some category", "Another label"]
			}
		]
	}`

	// Use httptest server as webhook URL (it will fail, but that's expected)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Simulate failure
		w.Write([]byte("Server error"))               //nolint:errcheck
	}))
	defer server.Close()

	req := httptest.NewRequest("POST", "/webhook?webhook_url="+server.URL, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")

	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/webhook", handler.HandleMinifluxWebhook)
	router.ServeHTTP(w, req)

	// Should still return OK even if the webhook call fails (error is logged)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Webhook processed successfully" {
		t.Errorf("Expected success message, got %v", response["message"])
	}
}
