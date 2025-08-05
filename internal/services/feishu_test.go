package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"miniflux-feishu/internal/models"
)

func TestFeishuService_FormatEntryMessage(t *testing.T) {
	service := NewFeishuService()

	// Test data based on the provided request
	publishedTime, _ := time.Parse(time.RFC3339, "2023-08-17T19:29:22Z")
	createdTime, _ := time.Parse(time.RFC3339, "2023-09-10T12:48:43.428196-07:00")

	entry := &models.WebhookEntry{
		ID:          231,
		UserID:      1,
		FeedID:      3,
		Status:      "unread",
		Hash:        "1163a93ef12741b558a3b86d7e975c4c1de0152f3439915ed185eb460e5718d7",
		Title:       "Example",
		URL:         "https://example.org/article",
		CommentsURL: "",
		Date:        publishedTime,
		CreatedAt:   createdTime,
		ChangedAt:   createdTime,
		Content:     "<p>Some HTML content</p>",
		ShareCode:   "",
		Starred:     false,
		ReadingTime: 1,
		Enclosures: []models.WebhookEnclosure{
			{
				ID:       158,
				UserID:   1,
				EntryID:  231,
				URL:      "https://example.org/podcast.mp3",
				MimeType: "audio/mpeg",
				Size:     63451045,
			},
		},
		Tags: []string{"Some category", "Another label"},
	}

	feed := &models.WebhookFeed{
		ID:        8,
		UserID:    1,
		FeedURL:   "https://example.org/feed.xml",
		SiteURL:   "https://example.org",
		Title:     "Example website",
		CheckedAt: createdTime,
	}

	message := service.formatEntryMessage(entry, feed)

	// Expected message structure
	expectedMessage := FeishuMessage{
		MsgType: "text",
		Content: FeishuTextContent{
			Title:   "[Example website] - Example",
			Content: "Some HTML content",
			URL:     "https://example.org/article",
		},
	}

	// Compare the actual result with expected
	if message.MsgType != expectedMessage.MsgType {
		t.Errorf("Expected MsgType %s, got %s", expectedMessage.MsgType, message.MsgType)
	}

	if message.Content.Title != expectedMessage.Content.Title {
		t.Errorf("Expected Title %s, got %s", expectedMessage.Content.Title, message.Content.Title)
	}

	if message.Content.Content != expectedMessage.Content.Content {
		t.Errorf("Expected Content %s, got %s", expectedMessage.Content.Content, message.Content.Content)
	}

	if message.Content.URL != expectedMessage.Content.URL {
		t.Errorf("Expected URL %s, got %s", expectedMessage.Content.URL, message.Content.URL)
	}

	// Print the actual JSON that would be sent to Feishu
	jsonBytes, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	t.Logf("Expected webhook body for Feishu:\n%s", string(jsonBytes))
}

func TestFeishuService_SendEntryToFeishu_Integration(t *testing.T) {
	// Create a mock HTTP server to capture the webhook request
	var capturedBody []byte
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()

		// Read the request body
		buf := make([]byte, r.ContentLength)
		_, err := r.Body.Read(buf)
		if err != nil && err.Error() != "EOF" {
			t.Errorf("Failed to read request body: %v", err)
		}
		capturedBody = buf

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	}))
	defer server.Close()

	service := NewFeishuService()

	// Test data based on the provided request
	publishedTime, _ := time.Parse(time.RFC3339, "2023-08-17T19:29:22Z")
	createdTime, _ := time.Parse(time.RFC3339, "2023-09-10T12:48:43.428196-07:00")

	entry := &models.WebhookEntry{
		ID:          231,
		UserID:      1,
		FeedID:      3,
		Status:      "unread",
		Hash:        "1163a93ef12741b558a3b86d7e975c4c1de0152f3439915ed185eb460e5718d7",
		Title:       "Example",
		URL:         "https://example.org/article",
		CommentsURL: "",
		Date:        publishedTime,
		CreatedAt:   createdTime,
		ChangedAt:   createdTime,
		Content:     "<p>Some HTML content</p>",
		ShareCode:   "",
		Starred:     false,
		ReadingTime: 1,
		Enclosures: []models.WebhookEnclosure{
			{
				ID:       158,
				UserID:   1,
				EntryID:  231,
				URL:      "https://example.org/podcast.mp3",
				MimeType: "audio/mpeg",
				Size:     63451045,
			},
		},
		Tags: []string{"Some category", "Another label"},
	}

	feed := &models.WebhookFeed{
		ID:        8,
		UserID:    1,
		FeedURL:   "https://example.org/feed.xml",
		SiteURL:   "https://example.org",
		Title:     "Example website",
		CheckedAt: createdTime,
	}

	// Send the entry to the mock server
	err := service.SendEntryToFeishu(entry, feed, server.URL)
	if err != nil {
		t.Fatalf("Failed to send entry to Feishu: %v", err)
	}

	// Verify the request headers
	if capturedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", capturedHeaders.Get("Content-Type"))
	}

	if capturedHeaders.Get("User-Agent") != "miniflux-feishu/1.0.0" {
		t.Errorf("Expected User-Agent miniflux-feishu/1.0.0, got %s", capturedHeaders.Get("User-Agent"))
	}

	// Parse and verify the captured body
	var capturedMessage FeishuMessage
	err = json.Unmarshal(capturedBody, &capturedMessage)
	if err != nil {
		t.Fatalf("Failed to unmarshal captured body: %v", err)
	}

	// Verify the message structure
	if capturedMessage.MsgType != "text" {
		t.Errorf("Expected MsgType text, got %s", capturedMessage.MsgType)
	}

	if capturedMessage.Content.Title != "[Example website] - Example" {
		t.Errorf("Expected Title '[Example website] - Example', got %s", capturedMessage.Content.Title)
	}

	if capturedMessage.Content.Content != "Some HTML content" {
		t.Errorf("Expected Content 'Some HTML content', got %s", capturedMessage.Content.Content)
	}

	if capturedMessage.Content.URL != "https://example.org/article" {
		t.Errorf("Expected URL https://example.org/article, got %s", capturedMessage.Content.URL)
	}

	// Print the actual webhook body that was sent
	t.Logf("Actual webhook body sent to Feishu:\n%s", string(capturedBody))
}

func TestFeishuService_StripHTML(t *testing.T) {
	service := NewFeishuService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple paragraph",
			input:    "<p>Some HTML content</p>",
			expected: "Some HTML content",
		},
		{
			name:     "multiple tags",
			input:    "<p>This is <strong>bold</strong> and <em>italic</em> text</p>",
			expected: "This is bold and italic text",
		},
		{
			name:     "line breaks",
			input:    "Line 1<br>Line 2<br/>Line 3<br />Line 4",
			expected: "Line 1\nLine 2\nLine 3\nLine 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
