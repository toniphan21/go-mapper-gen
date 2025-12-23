//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package multimappers

type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

type UserEntity struct {
	ID    string
	Name  string
	Email string
}
