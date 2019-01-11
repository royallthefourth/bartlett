package bartlett

import (
	"fmt"
	"log"
	"net/http"
)

func (b *Bartlett) Routes() (paths []string, handlers []func(http.ResponseWriter, *http.Request)) {
	paths = make([]string, len(b.Tables))
	handlers = make([]func(http.ResponseWriter, *http.Request), len(b.Tables))
	for i, t := range b.Tables {
		columns, err := b.Driver.GetColumns(b.DB, t)
		if err != nil {
			log.Println(err.Error())
		} else {
			t.columns = columns
		}
		paths[i] = fmt.Sprintf(`/%s`, t.Name)
		handlers[i] = b.handleRoute(t)
	}

	return paths, handlers
}

func (b Bartlett) handleRoute(table Table) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Content-Type`, `application/json`)

		if r.Method != `GET` {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		query, err := b.select_(table, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}

		rows, err := query.RunWith(b.DB).Query()
		defer rows.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}

		err = b.Driver.MarshalResults(rows, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}
	}
}
