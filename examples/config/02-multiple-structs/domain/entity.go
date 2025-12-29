
package domain

type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

type Address struct {
	ID          string
	HouseNumber string
	Street      string
	City        string
	State       string
	Country     string
}
