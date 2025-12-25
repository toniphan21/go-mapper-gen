
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *int16
	B int16
	C int
}

type Database struct {
	A pgtype.Int2
	B pgtype.Int2
	C pgtype.Int2
}
