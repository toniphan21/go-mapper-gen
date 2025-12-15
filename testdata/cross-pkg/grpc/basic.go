package grpc

type UserMessage struct {
	// The order of fields is shuffled to ensure the generator always follows
	// the order of the Target when mapping from Source to Target.

	Email string
	Id    string
	Age   int
	Name  string
}
