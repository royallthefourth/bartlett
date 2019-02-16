package bartlett

import (
	"errors"
	"fmt"
	"github.com/elgris/sqrl"
	"net/http"
	"strconv"
	"strings"
)

func (b Bartlett) buildUpdate(t Table, r *http.Request, userID interface{}, body []byte) (*sqrl.UpdateBuilder, error) {
	query := t.prepareUpdate(body, userID, sqrl.Update(t.Name))
	query, err := updateWhere(query, t, r)
	if err != nil {
		return query, err
	}
	query = updateOrder(query, t, r)
	query = updateLimit(query, r)

	if t.UserID != `` && userID != nil {
		query = query.Where(sqrl.Eq{t.UserID: userID})
	}

	return query, nil
}

func updateLimit(query *sqrl.UpdateBuilder, r *http.Request) *sqrl.UpdateBuilder {
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

func updateOrder(query *sqrl.UpdateBuilder, t Table, r *http.Request) *sqrl.UpdateBuilder {
	for _, col := range parseOrder(t, r) {
		query = query.OrderBy(fmt.Sprintf(`%s %s`, col.Column, strings.ToUpper(col.Direction)))
	}

	return query
}

func updateWhere(query *sqrl.UpdateBuilder, t Table, r *http.Request) (*sqrl.UpdateBuilder, error) {
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
					if parsedCond == `not.in` {
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
		err = errors.New(`UPDATE operations must have at least one WHERE clause`)
	}
	return query, err
}
