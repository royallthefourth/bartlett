package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/royallthefourth/bartlett"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func Routes(db *sql.DB, p bartlett.UserIDProvider, tables []bartlett.Table) (paths []string, handlers []func(http.ResponseWriter, *http.Request)) {
	paths = make([]string, len(tables))
	handlers = make([]func(http.ResponseWriter, *http.Request), len(tables))
	for i, t := range tables {
		paths[i] = fmt.Sprintf("/%s", t.Name)
		handlers[i] = handleRoute(t, db)
	}

	return paths, handlers
}

func handleRoute(table bartlett.Table, db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != `GET` {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		query := fmt.Sprintf(`SELECT * FROM %s`, table.Name)

		rows, err := db.Query(query)
		defer rows.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}

		err = sqlToJSON(rows, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(r.RequestURI + err.Error())
			return
		}
	}
}

// Adapted from https://stackoverflow.com/questions/42774467/how-to-convert-sql-rows-to-typed-json-in-golang
func sqlToJSON(rows *sql.Rows, w http.ResponseWriter) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf(`column error: %v`, err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf(`column type error: %v`, err)
	}

	types := make([]reflect.Type, len(columnTypes))
	for i, columnType := range columnTypes {
		scanType := columnType.ScanType()
		if scanType != nil {
			types[i] = scanType
		} else {
			switch strings.ToLower(columnType.DatabaseTypeName()) {
			case `null`:
				types[i] = nil
			case `integer`:
				types[i] = reflect.TypeOf(int(0))
			case `float`:
				types[i] = reflect.TypeOf(float64(0))
			case `blob`:
				types[i] = reflect.TypeOf([]byte{})
			case `text`:
				types[i] = reflect.TypeOf(``)
			case `timestamp`:
				types[i] = reflect.TypeOf(time.Now())
			case `datetime`:
				types[i] = reflect.TypeOf(time.Now())
			default:
				return fmt.Errorf(`scantype is nil for column %v`, columnType)
			}
		}
	}

	values := make([]interface{}, len(columnTypes))
	data := make(map[string]interface{})

	_, err = w.Write([]byte{'['})
	if err != nil {
		return fmt.Errorf(`failed to write opening bracket: %s`, err)
	}

	count := 0
	for rows.Next() {
		if count > 0 {
			_, err = w.Write([]byte{','})
			if err != nil {
				return fmt.Errorf(`failed to write comma: %s`, err)
			}
		}
		count++
		for i := range values {
			values[i] = reflect.New(types[i]).Interface()
		}
		err = rows.Scan(values...)
		if err != nil {
			return fmt.Errorf(`failed to scan values: %v`, err)
		}
		for i, v := range values {
			data[columns[i]] = v
		}

		jsonRow, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf(`failed to marshal to json: %s`, err)
		}

		_, err = w.Write(jsonRow)
		if err != nil {
			return fmt.Errorf(`failed to write closing bracket: %s`, err)
		}
	}

	_, err = w.Write([]byte{']'})
	if err != nil {
		return fmt.Errorf(`failed to write closing bracket: %s`, err)
	}

	return err
}