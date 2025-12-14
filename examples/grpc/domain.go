package grpc

import _ "github.com/toniphan21/go-mapper-gen"

//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

type User struct {
	ID       string
	Email    string
	Password string
}

type UserMessage struct {
	Id       string
	Email    string
	Password string
}
