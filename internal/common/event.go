package common

import (
	"time"

	"github.com/google/uuid"
)

const (
	HistoryTopic      = "history"
	NotificationTopic = "notification"
	EmailSent         = "email_sent"
)

type Event struct {
	ID   uuid.UUID   `json:"id"`
	Type string      `json:"type"`
	Time time.Time   `json:"time"`
	Data interface{} `json:"data"`
	User uuid.UUID   `json:"user"`
}

type EmailData struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
