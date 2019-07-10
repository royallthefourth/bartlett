package bartlett

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
)

type whereCond struct {
	Column    string // necessary due to the way Squirrel handles IN
	Condition string
	Value     string
}

func buildConds(t Table, r *http.Request) []whereCond {
	i := 0
	columns := make([]string, len(r.URL.Query()))
	for k := range r.URL.Query() {
		columns[i] = k
		i++
	}
	columns = t.validReadColumns(columns)
	conds := make([]whereCond, 0)

	for column, values := range r.URL.Query() {
		if sliceContains(columns, column) {
			for _, rawCond := range values {
				parsedCond, val := parseWhereCond(rawCond)
				var cond string
				if parsedCond == `in` || parsedCond == `not.in` {
					conds = append(conds, whereCond{column, parsedCond, val})
				} else {
					cond = urlToWhereCond(column, parsedCond)
					val := rectifyArg(cond, val)
					conds = append(conds, whereCond{column, cond, val})
				}
			}
		}
	}

	return conds
}

func parseSimpleWhereCond(rawCond string) (cond, val string) {
	parts := strings.Split(rawCond, `.`)
	if parts[0] == `not` {
		cond = fmt.Sprintf(`%s.%s`, parts[0], parts[1])
	} else {
		cond = parts[0]
	}

	val = strings.Replace(rawCond, cond+`.`, ``, 1)
	return cond, val
}

func parseWhereCond(rawCond string) (cond, val string) {
	if rawCond[0:2] == `or` {
		return "", ""
	} else {
		return parseSimpleWhereCond(rawCond)
	}
}

func rectifyArg(cond, val string) string {
	if strings.Contains(cond, `LIKE ?`) {
		val = strings.Replace(val, `*`, `%`, -1)
	}
	return val
}

func urlToWhereCond(column, condition string) string {
	switch condition {
	case `eq`:
		return fmt.Sprintf(`%s = ?`, column)
	case `not.eq`:
		return fmt.Sprintf(`%s != ?`, column)
	case `neq`:
		return fmt.Sprintf(`%s != ?`, column)
	case `not.neq`:
		return fmt.Sprintf(`%s = ?`, column)
	case `gt`:
		return fmt.Sprintf(`%s > ?`, column)
	case `not.gt`:
		return fmt.Sprintf(`%s <= ?`, column)
	case `gte`:
		return fmt.Sprintf(`%s >= ?`, column)
	case `not.gte`:
		return fmt.Sprintf(`%s < ?`, column)
	case `lt`:
		return fmt.Sprintf(`%s < ?`, column)
	case `not.lt`:
		return fmt.Sprintf(`%s >= ?`, column)
	case `lte`:
		return fmt.Sprintf(`%s <= ?`, column)
	case `not.lte`:
		return fmt.Sprintf(`%s > ?`, column)
	case `like`:
		return fmt.Sprintf(`%s LIKE ?`, column)
	case `not.like`:
		return fmt.Sprintf(`%s NOT LIKE ?`, column)
	case `is`:
		return fmt.Sprintf(`%s IS ?`, column)
	case `not.is`:
		return fmt.Sprintf(`%s IS NOT ?`, column)
	default:
		return ``
	}
}

func whereIn(rawVal string) []string {
	r := csv.NewReader(strings.NewReader(strings.TrimPrefix(strings.TrimSuffix(rawVal, `)`), `(`)))
	vals, _ := r.Read()
	return vals
}
