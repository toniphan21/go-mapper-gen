
package example

import "time"

type User struct {
	Name      string
	ID        string
	Age       int
	Email     string
	UpdatedAt time.Time
	CreatedAt time.Time
}
