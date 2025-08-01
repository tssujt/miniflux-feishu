package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux-feishu/internal/models"
)

type FeishuService struct {
	client *http.Client
}

type FeishuMessage struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"`
}

type FeishuRichTextContent struct {
	RichText *FeishuRichText `json:"rich_text"`
}

type FeishuRichText struct {
	Elements [][]FeishuElement `json:"elements"`
}

type FeishuElement struct {
	Tag   string   `json:"tag"`
	Text  string   `json:"text,omitempty"`
	Href  string   `json:"href,omitempty"`
	Style []string `json:"style,omitempty"`
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
	elements := [][]FeishuElement{
		{
			{Tag: "text", Text: fmt.Sprintf("📰 %s", feed.Title), Style: []string{"bold"}},
		},
		{
			{Tag: "text", Text: ""},
		},
		{
			{Tag: "a", Text: entry.Title, Href: entry.URL},
		},
	}

	if entry.Content != "" {
		content := s.stripHTML(entry.Content)
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		elements = append(elements, []FeishuElement{
			{Tag: "text", Text: ""},
		})
		elements = append(elements, []FeishuElement{
			{Tag: "text", Text: content},
		})
	}

	if len(entry.Tags) > 0 {
		tagText := "🏷️ " + strings.Join(entry.Tags, ", ")
		elements = append(elements, []FeishuElement{
			{Tag: "text", Text: ""},
		})
		elements = append(elements, []FeishuElement{
			{Tag: "text", Text: tagText, Style: []string{"italic"}},
		})
	}

	publishedTime := entry.Date.Format("2006-01-02 15:04")
	elements = append(elements, []FeishuElement{
		{Tag: "text", Text: ""},
	})
	elements = append(elements, []FeishuElement{
		{Tag: "text", Text: fmt.Sprintf("🕐 %s", publishedTime), Style: []string{"italic"}},
	})

	return FeishuMessage{
		MsgType: "rich_text",
		Content: FeishuRichTextContent{
			RichText: &FeishuRichText{
				Elements: elements,
			},
		},
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBody bytes.Buffer
		respBody.ReadFrom(resp.Body)
		return fmt.Errorf("feishu API returned status %d: %s", resp.StatusCode, respBody.String())
	}

	return nil
}
