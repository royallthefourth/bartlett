package common

import (
	"database/sql"
	"fmt"
	"github.com/royallthefourth/bartlett"
	"net/http"
)

func Routes(db *sql.DB, p bartlett.UserIDProvider, makeRoute func(table bartlett.Table, db *sql.DB) func(http.ResponseWriter, *http.Request), tables []bartlett.Table) (paths []string, handlers []func(http.ResponseWriter, *http.Request)) {
	paths = make([]string, len(tables))
	handlers = make([]func(http.ResponseWriter, *http.Request), len(tables))
	for i, t := range tables {
		paths[i] = fmt.Sprintf("/%s", t.Name)
		handlers[i] = makeRoute(t, db)
	}

	return paths, handlers
}
