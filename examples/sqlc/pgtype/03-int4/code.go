
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *int32
	B int32
	C int
}

type Database struct {
	A pgtype.Int4
	B pgtype.Int4
	C pgtype.Int4
}
