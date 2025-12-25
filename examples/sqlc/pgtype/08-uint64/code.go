
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *uint64
	B uint64
	C int
}

type Database struct {
	A pgtype.Uint64
	B pgtype.Uint64
	C pgtype.Uint64
}
