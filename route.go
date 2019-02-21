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
func (b *Bartlett) Routes() (paths []string, handlers []http.HandlerFunc) {
	paths = make([]string, len(b.Tables))
	handlers = make([]http.HandlerFunc, len(b.Tables))
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

		switch r.Method {
		case http.MethodGet:
			b.handleGet(t, w, r)
		case http.MethodPost:
			b.handlePost(t, w, r)
		case http.MethodDelete:
			b.handleDelete(t, w, r)
		case http.MethodPatch:
			b.handlePatch(t, w, r)
		default:
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
	}
}

func (b Bartlett) handleDelete(t Table, w http.ResponseWriter, r *http.Request) {
	if !t.Writable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "Table %s is read-only"}`, t.Name)))
		return
	}

	query, err := b.buildDelete(t, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	rows, err := query.RunWith(b.DB).Query()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}
	defer rows.Close()

	err = b.Driver.MarshalResults(rows, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}
}

func (b Bartlett) handleGet(t Table, w http.ResponseWriter, r *http.Request) {
	query, err := b.buildSelect(t, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	rows, err := query.RunWith(b.DB).Query()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}
	defer rows.Close()

	err = b.Driver.MarshalResults(rows, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}
}

func (b Bartlett) handlePatch(t Table, w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	status, userID, err := b.validateWrite(t, r, body)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	query, err := b.buildUpdate(t, r, userID, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	_, err = query.RunWith(b.DB).Exec()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (b Bartlett) handlePost(t Table, w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	status, userID, err := b.validateWrite(t, r, body)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	tx, err := b.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	var n uint
	_, err = jsonparser.ArrayEach(body, func(row []byte, dataType jsonparser.ValueType, offset int, err error) {
		query := t.prepareInsert(row, userID)
		_, err = query.RunWith(tx).Exec()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
			return
		}
		n++
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s in %s"}`, err.Error(), body)))
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf(`{"inserts": %d}`, n)))
}

func (b Bartlett) validateWrite(t Table, r *http.Request, body []byte) (status int, userID interface{}, err error) {
	status = http.StatusOK

	if !t.Writable {
		status = http.StatusMethodNotAllowed
		err = fmt.Errorf(`table %s is read-only`, t.Name)
		return status, nil, err
	}

	if !json.Valid(body) {
		status = http.StatusBadRequest
		err = fmt.Errorf(`JSON data not valid`)
		return status, userID, err
	}

	if r.Method == http.MethodPost && rune(body[0]) != '[' { // Inserts are arrays.
		status = http.StatusBadRequest
		err = fmt.Errorf(`JSON data should be an array`)
		return status, nil, err
	} else if r.Method == http.MethodPatch && rune(body[0]) != '{' { // Updates are single value.
		status = http.StatusBadRequest
		err = fmt.Errorf(`JSON data should be an object`)
		return status, nil, err
	}

	if t.UserID != `` {
		userID, err = b.Users(r)
		if err != nil || userID == nil {
			status = http.StatusForbidden
			err = fmt.Errorf(`failed to generate userID: %s`, err.Error())
			return status, nil, err
		}
	} else {
		userID = 0
	}

	return status, userID, err
}
