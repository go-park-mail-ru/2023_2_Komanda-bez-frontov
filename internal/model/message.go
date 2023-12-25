package model

import "time"

type CheckUnreadMessages struct {
	Count int `json:"unread"`
}

type Message struct {
	ID       *int64    `json:"id"`
	SenderID int64     `json:"sender_id" validate:"required"`
	Text     string    `json:"text" validate:"required"`
	SendAt   time.Time `json:"send_at"`
	IsRead   bool      `json:"is_read"`
}

type MessageSave struct {
	SenderID   int64     `json:"sender_id" validate:"required"`
	ReceiverID int64     `json:"receiver_id" validate:"required"`
	Text       string    `json:"text" validate:"required"`
	SendAt     time.Time `json:"send_at"`
	// FormID		int64		`json:"form_id"`
}

type Chat struct {
	User *UserGet `json:"user"`
	// Form		*FormMessage `json:"form"`
	Messages []*Message `json:"messages"`
}

type ChatList struct {
	CollectionResponse
	Chats []*Chat `json:"chats" validate:"required"`
}
