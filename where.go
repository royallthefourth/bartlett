package bartlett

import (
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

func urlToWhereCond(column, condition string) string {
	switch condition {
	case `eq`:
		return fmt.Sprintf(`%s = ?`, column)
	case `not.eq`:
		return fmt.Sprintf(`%s != ?`, column)
	default:
		return ``
	}
}
