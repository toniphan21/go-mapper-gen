package domain

import "time"

type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}
