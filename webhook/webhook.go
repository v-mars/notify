package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/types"
	"net/http"
	"time"
)

// Webhook represents a webhook notification configuration
type Webhook struct {
	types.Webhook
}

// Message represents the message structure sent to webhook endpoint
type Message struct {
	To      []string `json:"to"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
}

// Result represents the response from webhook endpoint
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// NewWebhook creates a new webhook sender
func NewWebhook(url string, timeout time.Duration, headers map[string]string) *Webhook {
	return &Webhook{
		Webhook: types.Webhook{
			URL:     url,
			Timeout: timeout,
			Headers: headers,
		},
	}
}

// Send sends notification via webhook
func (w *Webhook) Send(to []string, title string, content string) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeWebhook,
		ChannelMsgID: nil,
		Success:      false,
		MessageID:    "",
		SendTime:     time.Now(),
		Error:        nil,
		CostMs:       0,
	}
	defer func() {
		sendResult.CostMs = time.Now().Sub(sendResult.SendTime).Milliseconds()
		sendResult.ChannelMsgID = result.PtrOf(fmt.Sprintf("%d", time.Now().UnixNano()))
		sendResult.Success = err == nil
		sendResult.MessageID = *sendResult.ChannelMsgID
		if err != nil {
			sendResult.Error = result.PtrOf(err.Error())
		}
	}()
	message := Message{
		To:      to,
		Title:   title,
		Content: content,
	}

	client := &http.Client{
		Timeout: w.Timeout,
	}

	headers := w.Headers
	if headers == nil {
		headers = make(map[string]string)
	}

	// Add default content type if not present
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	respData, err := notify.JSONPost(http.MethodPost, w.URL, message, client, headers)
	if err != nil {
		return sendResult, fmt.Errorf("failed to send webhook notification: %w", err)
	}

	result := Result{}
	if err = json.Unmarshal(respData, &result); err != nil {
		// If we can't parse the response, we assume success if we got a response
		return sendResult, nil
	}

	if !result.Success {
		return sendResult, fmt.Errorf("webhook endpoint returned failure: %s", result.Message)
	}

	return sendResult, nil
}

const NotifyTypeWebhook = "webhook"

// ChannelType returns the channel type
func (w *Webhook) ChannelType() string {
	return NotifyTypeWebhook
}
