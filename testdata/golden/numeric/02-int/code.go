//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package numeric

type Typed int
type Alias = int

type Target struct {
	A int
	B int
	C int
	D int
	E int
	F int
	G int
	H int
	I int
	J int
	K int
	L int
	M int
	N int
	O int
}

type Source struct {
	A byte
	B uint8
	C uint16
	D uint32
	E uint64
	F int8
	G int16
	H int32
	I int64
	J uint
	K int
	L float32
	M float64	
	N Typed
	O Alias
}
