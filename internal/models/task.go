package models

import "time"

// Статусы задачи по жизненному циклу: разместили -> кто-то взял в работу -> закрыли.
const (
	TaskStatusOpen       = "open"
	TaskStatusInProgress = "in_progress"
	TaskStatusClosed     = "closed"
)

type Task struct {
	ID           int        `json:"id"`
	CustomerID   int        `json:"customer_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Requirements string     `json:"requirements"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}
