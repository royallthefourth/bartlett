package common

import (
	"github.com/elgris/sqrl"
	"github.com/royallthefourth/bartlett"
	"net/http"
)

func Select(table bartlett.Table, users bartlett.UserIDProvider, r *http.Request) (*sqrl.SelectBuilder, error) {
	query := sqrl.Select(`*`).From(table.Name)

	if table.UserID != `` {
		userID, err := users(r)
		if err != nil {
			return query, err
		}
		query = query.Where(sqrl.Eq{table.UserID: userID})
	}

	return query, nil
}
