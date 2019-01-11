package bartlett

import (
	"fmt"
	"github.com/elgris/sqrl"
	"net/http"
	"strings"
)

func (b Bartlett) select_(t Table, r *http.Request) (*sqrl.SelectBuilder, error) {
	query := selectColumns(t, r).From(t.Name)
	query = selectOrder(query, t, r)

	if t.UserID != `` {
		userID, err := b.Users(r)
		if err != nil {
			return query, err
		}
		query = query.Where(sqrl.Eq{t.UserID: userID})
	}

	return query, nil
}

type orderSpec struct {
	Column    string
	Direction string
}

func selectOrder(query *sqrl.SelectBuilder, t Table, r *http.Request) *sqrl.SelectBuilder {
	for _, col := range parseOrder(t, r) {
		query = query.OrderBy(fmt.Sprintf(`%s %s`, col.Column, strings.ToUpper(col.Direction)))
	}

	return query
}

func parseOrder(t Table, r *http.Request) []orderSpec {
	out := make([]orderSpec, 0)

	if len(r.URL.Query()[`order`]) > 0 {
		orderCols := strings.Split(r.URL.Query()[`order`][0], `,`)
		order := orderSpec{}
		for _, col := range orderCols {
			order.Direction = `desc`
			if strings.Contains(col, `.`) {
				pair := strings.Split(col, `.`)
				order.Column = pair[0]
				if strings.ToLower(pair[1]) == `asc` {
					order.Direction = `asc`
				}
			} else {
				order.Column = col
			}
			if sliceContains(t.columns, order.Column) {
				out = append(out, order) // Omit anything not in the table spec
			}
		}
	}

	return out
}

func selectColumns(t Table, r *http.Request) *sqrl.SelectBuilder {
	var query *sqrl.SelectBuilder
	columns := parseColumns(t, r)
	if len(columns) > 0 {
		query = sqrl.Select(columns[0])
		query = query.Columns(columns[1:]...)
	} else {
		query = sqrl.Select(`*`)
	}

	return query
}

func parseColumns(t Table, r *http.Request) []string {
	out := make([]string, 0)

	if len(r.URL.Query()[`select`]) > 0 {
		requestColumns := strings.Split(r.URL.Query()[`select`][0], `,`) // Get the first `select` var and forget about any others
		for _, col := range requestColumns {
			if sliceContains(t.columns, col) {
				out = append(out, col)
			}
		}
	}

	return out
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}

	return false
}
