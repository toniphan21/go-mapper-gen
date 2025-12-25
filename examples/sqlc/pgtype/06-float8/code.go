
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *float64
	B float64
	C int
}

type Database struct {
	A pgtype.Float8
	B pgtype.Float8
	C pgtype.Float8
}
