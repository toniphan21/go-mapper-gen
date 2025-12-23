//go:generate go run github.com/toniphan21/go-mapper-gen/cmd/generator

package numeric

type Typed uint
type Alias = uint

type Target struct {
	A uint
	B uint
	C uint
	D uint
	E uint
	F uint
	G uint
	H uint
	I uint
	J uint
	K uint
	L uint
	M uint
	N uint
	O uint
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
