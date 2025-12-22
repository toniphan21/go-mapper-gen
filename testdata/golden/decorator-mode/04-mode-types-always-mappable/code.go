//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package decorator

type User struct {
	ID   string
	Name string
}

type UserEntity struct {
	ID    string
	Email string
}

type UserMessage struct {
	ID    string
	Name  string
}
