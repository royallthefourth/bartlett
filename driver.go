package bartlett

import (
	"database/sql"
	"net/http"
)

// The Driver interface contains database-specific code, which I'm trying to keep to a minimum.
// Implement a column-identifying function and a result marshaling function for your database of choice.
type Driver interface {
	GetColumns(db *sql.DB, t Table) ([]string, error)
	MarshalResults(rows *sql.Rows, w http.ResponseWriter) error
}
