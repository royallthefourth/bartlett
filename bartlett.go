package bartlett

import (
	"database/sql"
	"net/http"
)

// A Table represents a table in the database.
// Name is required.
type Table struct {
	Name     string
	ReadOnly bool
	UserID   string
}

// A UserIDProvider is a function that is able to use an incoming request to produce a user ID.
type UserIDProvider func(r *http.Request) (interface{}, error)

type Bartlett interface {
	Routes(db *sql.DB, p UserIDProvider, tables []Table) (paths []string, handlers []func(http.ResponseWriter, *http.Request))
}
