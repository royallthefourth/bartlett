package bartlett

import (
	"github.com/elgris/sqrl"
	"net/http"
	"strings"
)

func (b Bartlett) select_(t Table, r *http.Request) (*sqrl.SelectBuilder, error) {
	var query *sqrl.SelectBuilder
	columns := parseColumns(t, r)
	if len(columns) > 0 {
		query = sqrl.Select(columns[0])
		query = query.Columns(columns[1:]...)
	} else {
		query = sqrl.Select(`*`)
	}

	query = query.From(t.Name)

	if t.UserID != `` {
		userID, err := b.Users(r)
		if err != nil {
			return query, err
		}
		query = query.Where(sqrl.Eq{t.UserID: userID})
	}

	return query, nil
}

func parseColumns(t Table, r *http.Request) []string {
	out := make([]string, 0)

	if len(r.URL.Query()[`select`]) > 0 {
		requestColumns := strings.Split(r.URL.Query()[`select`][0], `,`) // Get the first `select` var and forget about any others
		for _, col := range requestColumns {
			if contains(t.columns, col) {
				out = append(out, col)
			}
		}
	}

	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}

	return false
}
