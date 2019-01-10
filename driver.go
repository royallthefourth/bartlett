package bartlett

import (
	"database/sql"
	"net/http"
)

type Driver interface {
	GetColumns(db *sql.DB, t Table) ([]string, error)
	MarshalResults(rows *sql.Rows, w http.ResponseWriter) error
}
