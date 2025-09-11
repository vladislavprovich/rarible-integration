package mtproto

import (
	"context"
	"time"
)

// Message represents a Telegram message
type Message struct {
	ID       int64     `json:"id"`
	Text     string    `json:"text"`
	UserID   int64     `json:"user_id"`
	ChatID   int64     `json:"chat_id"`
	SentAt   time.Time `json:"sent_at"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ChatID   int64                  `json:"chat_id"`
	Text     string                 `json:"text"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageResponse represents a response after sending a message
type SendMessageResponse struct {
	MessageID int64     `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
	Success   bool      `json:"success"`
}

// GetMessagesRequest represents a request to get messages
type GetMessagesRequest struct {
	ChatID int64 `json:"chat_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// GetMessagesResponse represents a response containing messages
type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	HasMore  bool      `json:"has_more"`
}

// Client defines the interface for MTProto client operations
type Client interface {
	// SendMessage sends a message to the specified chat
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	
	// GetMessages retrieves messages from the specified chat
	GetMessages(ctx context.Context, req *GetMessagesRequest) (*GetMessagesResponse, error)
	
	// Connect establishes connection to Telegram
	Connect(ctx context.Context) error
	
	// Disconnect closes connection to Telegram
	Disconnect(ctx context.Context) error
	
	// IsConnected returns connection status
	IsConnected() bool
}