
package example

import "github.com/jackc/pgx/v5/pgtype"

type Domain struct {
	A *bool
	B bool
}

type Database struct {
	A pgtype.Bool
	B pgtype.Bool
}
