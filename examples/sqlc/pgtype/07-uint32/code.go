
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *uint32
	B uint32
	C int
}

type Database struct {
	A pgtype.Uint32
	B pgtype.Uint32
	C pgtype.Uint32
}
