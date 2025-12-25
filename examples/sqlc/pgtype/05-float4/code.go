
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *float32
	B float32
	C int
}

type Database struct {
	A pgtype.Float4
	B pgtype.Float4
	C pgtype.Float4
}
