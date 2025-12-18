package domain

type MissingFieldTarget struct {
	ID        string
	Gender    string
	LastName  string
	Email     string
	Age       int
	FirstName string
}

type MissingFieldSource struct {
	ID    string
	Age   int
	Email string
}
