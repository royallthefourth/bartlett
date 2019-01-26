package bartlett

import (
	"errors"
	"fmt"
	"github.com/elgris/sqrl"
	"net/http"
	"strconv"
	"strings"
)

func (b Bartlett) buildDelete(t Table, r *http.Request) (*sqrl.DeleteBuilder, error) {
	query, err := deleteWhere(sqrl.Delete(t.Name), t, r)
	if err != nil {
		return query, err
	}
	query = deleteOrder(query, t, r)
	query = deleteLimit(query, r)

	if t.UserID != `` {
		userID, err := b.Users(r)
		if err != nil {
			return query, err
		}
		query = query.Where(sqrl.Eq{t.UserID: userID})
	}

	return query, nil
}

func deleteLimit(query *sqrl.DeleteBuilder, r *http.Request) *sqrl.DeleteBuilder {
	var (
		err   error
		limit int
	)
	if len(r.URL.Query()[`limit`]) > 0 {
		limit, err = strconv.Atoi(r.URL.Query()[`limit`][0])
		if err != nil || limit < 0 {
			limit = 0
		}
	}
	if limit > 0 {
		query = query.Limit(uint64(limit))
	}

	return query
}

func deleteOrder(query *sqrl.DeleteBuilder, t Table, r *http.Request) *sqrl.DeleteBuilder {
	for _, col := range parseOrder(t, r) {
		query = query.OrderBy(fmt.Sprintf(`%s %s`, col.Column, strings.ToUpper(col.Direction)))
	}

	return query
}

func deleteWhere(query *sqrl.DeleteBuilder, t Table, r *http.Request) (*sqrl.DeleteBuilder, error) {
	var err error = nil
	whereClauses := 0
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
					if cond == `not.in` {
						query = query.Where(sqrl.NotEq{column: whereIn(val)})
					} else {
						query = query.Where(sqrl.Eq{column: whereIn(val)})
					}
				} else {
					cond = urlToWhereCond(column, parsedCond)
					sqlCond, val := rectifyArg(cond, val)
					query = query.Where(sqlCond, val)
				}
				whereClauses++
			}
		}
	}

	if whereClauses == 0 {
		err = errors.New(`DELETE operations must have at least one WHERE clause`)
	}
	return query, err
}
