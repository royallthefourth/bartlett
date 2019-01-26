package bartlett

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"net/http"
)

// Routes generates all of the URLs and handlers for the tables specified in Bartlett.
// Iterate this output to feed it into your web server, prefix or otherwise alter the route names,
// and add filtering to the handler functions.
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

func (b Bartlett) handleRoute(t Table) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Content-Type`, `application/json`)

		if r.Method == http.MethodGet {
			b.handleGet(t, w, r)
		} else if r.Method == http.MethodPost {
			b.handlePost(t, w, r)
		} else if r.Method == http.MethodDelete {
			b.handleDelete(t, w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
	}
}

func (b Bartlett) handleGet(t Table, w http.ResponseWriter, r *http.Request) {
	query, err := b.buildSelect(t, r)
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

func (b Bartlett) handleDelete(t Table, w http.ResponseWriter, r *http.Request) {
	if !t.Writable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println(r.RequestURI + ` Invalid insert attempt to read-only ` + t.Name)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "Table %s is read-only"}`, t.Name)))
		return
	}

	query, err := b.buildDelete(t, r)
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

func (b Bartlett) handlePost(t Table, w http.ResponseWriter, r *http.Request) {
	if !t.Writable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println(r.RequestURI + ` Invalid insert attempt to read-only ` + t.Name)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "Table %s is read-only"}`, t.Name)))
		return
	}

	var (
		userID interface{}
		err    error
	)
	if len(t.UserID) > 0 {
		userID, err = b.Users(r)
		if err != nil || userID == nil {
			w.WriteHeader(http.StatusForbidden)
			log.Println(r.RequestURI + err.Error())
			return
		}
	} else {
		userID = 0
	}

	rawBody, err := ioutil.ReadAll(r.Body)
	if rune(rawBody[0]) != '[' || !json.Valid(rawBody) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(r.RequestURI + ` Invalid JSON data post`)
		_, _ = w.Write([]byte(`{"error": "JSON data should be an array"}`))
		return
	}

	tx, err := b.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(r.RequestURI + err.Error())
		return
	}

	n, err := jsonparser.ArrayEach(rawBody, func(row []byte, dataType jsonparser.ValueType, offset int, err error) {
		query := t.prepareInsert(row, userID)
		_, err = query.RunWith(tx).Exec()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}

	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(r.RequestURI + err.Error())
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(r.RequestURI + err.Error())
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf(`{"inserts": %d}`, n)))
}
