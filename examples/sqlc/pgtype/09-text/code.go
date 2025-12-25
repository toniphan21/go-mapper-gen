
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *string
	B string
}

type Database struct {
	A pgtype.Text
	B pgtype.Text
}
