//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package rest

type UserResponse struct {
	ID    string
	Name  string
	Email string
	Age   int
}

type AddressResponse struct {
	ID          string
	HouseNumber string
	Street      string
	City        string
	State       string
	Country     string
}
