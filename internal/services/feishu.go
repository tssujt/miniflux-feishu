package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"miniflux-feishu/internal/models"
)

type FeishuService struct {
	client *http.Client
}

type FeishuMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

func NewFeishuService() *FeishuService {
	return &FeishuService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *FeishuService) SendEntryToFeishu(entry *models.WebhookEntry, feed *models.WebhookFeed, webhookURL string) error {
	message := s.formatEntryMessage(entry, feed)
	return s.sendMessage(message, webhookURL)
}

func (s *FeishuService) formatEntryMessage(entry *models.WebhookEntry, feed *models.WebhookFeed) FeishuMessage {
	content := ""
	if entry.Content != "" {
		content = s.stripHTML(entry.Content)
		if len(content) > 300 {
			content = content[:300] + "..."
		}
	}

	title := fmt.Sprintf("[%s] - %s", feed.Title, entry.Title)

	return FeishuMessage{
		Title:   title,
		Content: content,
		Link:    entry.URL,
	}
}

func (s *FeishuService) stripHTML(html string) string {
	result := html
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	result = strings.ReplaceAll(result, "<p>", "\n")
	result = strings.ReplaceAll(result, "</p>", "")

	inTag := false
	var cleaned strings.Builder
	for _, char := range result {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			cleaned.WriteRune(char)
		}
	}

	text := cleaned.String()
	text = strings.ReplaceAll(text, "\n\n", "\n")
	return strings.TrimSpace(text)
}

func (s *FeishuService) sendMessage(message FeishuMessage, webhookURL string) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "miniflux-feishu/1.0.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var respBody bytes.Buffer
		if _, err := respBody.ReadFrom(resp.Body); err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("feishu API returned status %d: %s", resp.StatusCode, respBody.String())
	}

	return nil
}
