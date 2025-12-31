
package awesome

// you can add more converters and use your own binary to generate mapper now.
//go:generate go run ./cmd/generator

type Target struct {
	ID   string
	Name string
}

type Source struct {
	ID   string
	Name string
}
