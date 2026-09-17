package models

import "time"

// Роли пользователей на платформе. Заказчик размещает задачи,
// исполнитель на них откликается.
const (
	RoleCustomer = "customer"
	RoleExecutor = "executor"
)

type User struct {
	ID           int       `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // хэш никогда не отдаём наружу через API
	Role         string    `json:"role"`
	Competencies string    `json:"competencies"`
	CreatedAt    time.Time `json:"created_at"`
}
