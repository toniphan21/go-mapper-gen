
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *int64
	B int64
	C int
}

type Database struct {
	A pgtype.Int8
	B pgtype.Int8
	C pgtype.Int8
}
