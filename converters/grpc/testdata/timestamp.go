package grpc

import "time"

type User struct {
	ID        string
	Age       int32
	LastName  string
	FirstName string
	Email     string
	UpdatedAt time.Time
	CreatedAt time.Time
}
