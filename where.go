package bartlett

import (
	"encoding/csv"
	"fmt"
	"strings"
)

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

func rectifyArg(cond, val string) (string, string) {
	if strings.Contains(cond, `LIKE ?`) {
		val = strings.Replace(val, `*`, `%`, -1)
	}
	return cond, val
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
