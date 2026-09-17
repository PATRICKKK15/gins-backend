package models

import "time"

// Статусы отклика: ждёт решения заказчика, принят или отклонён.
const (
	ResponseStatusPending  = "pending"
	ResponseStatusAccepted = "accepted"
	ResponseStatusRejected = "rejected"
)

type Response struct {
	ID         int       `json:"id"`
	TaskID     int       `json:"task_id"`
	ExecutorID int       `json:"executor_id"`
	Motivation string    `json:"motivation"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
