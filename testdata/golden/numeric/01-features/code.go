//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package numeric

type Target struct {
	A int
	B int
	C *int
	D *int
}

type Source struct {
	A uint
	B *uint
	C uint
	D *uint
}
