package bartlett

import (
	"github.com/elgris/sqrl"
	"net/http"
)

func (b Bartlett) select_(table Table, r *http.Request) (*sqrl.SelectBuilder, error) {
	query := sqrl.Select(`*`).From(table.Name)

	if table.UserID != `` {
		userID, err := b.Users(r)
		if err != nil {
			return query, err
		}
		query = query.Where(sqrl.Eq{table.UserID: userID})
	}

	return query, nil
}
