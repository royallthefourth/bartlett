package bartlett

import (
	"fmt"
	"github.com/elgris/sqrl"
	"net/http"
	"strconv"
	"strings"
)

func (b Bartlett) buildSelect(t Table, r *http.Request) (*sqrl.SelectBuilder, error) {
	query := selectColumns(t, r).From(t.Name)
	query = selectWhere(query, t, r)
	query = selectOrder(query, t, r)
	query = selectLimit(query, r)

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
	var out []orderSpec

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
	var out []string

	if len(r.URL.Query()[`select`]) > 0 {
		requestColumns := strings.Split(r.URL.Query()[`select`][0], `,`) // Get the first `select` var and forget about any others.
		out = t.validReadColumns(requestColumns)
	}

	return out
}

func selectLimit(query *sqrl.SelectBuilder, r *http.Request) *sqrl.SelectBuilder {
	var (
		err    error
		limit  int
		offset int
	)
	if len(r.URL.Query()[`limit`]) > 0 {
		limit, err = strconv.Atoi(r.URL.Query()[`limit`][0])
		if err != nil || limit < 0 {
			limit = 0
		}
	}
	if limit > 0 && len(r.URL.Query()[`offset`]) > 0 {
		offset, err = strconv.Atoi(r.URL.Query()[`offset`][0])
		if err != nil || offset < 0 {
			offset = 0
		}
	}
	if limit > 0 {
		query = query.Limit(uint64(limit)).Offset(uint64(offset))
	}

	return query
}

func selectWhere(query *sqrl.SelectBuilder, t Table, r *http.Request) *sqrl.SelectBuilder {
	i := 0
	columns := make([]string, len(r.URL.Query()))
	for k := range r.URL.Query() {
		columns[i] = k
		i++
	}
	columns = t.validReadColumns(columns)

	for column, values := range r.URL.Query() {
		if sliceContains(columns, column) {
			for _, rawCond := range values {
				parsedCond, val := parseSimpleWhereCond(rawCond)
				var cond string
				if parsedCond == `in` || parsedCond == `not.in` {
					// TODO learn how to count args from IN queries
					// TODO convert IN to sqrl.Eq
				} else {
					cond = urlToWhereCond(column, parsedCond)
					sqlCond, val := rectifyArg(cond, val)
					query = query.Where(sqlCond, val)
				}
			}
		}
	}

	return query
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}

	return false
}
