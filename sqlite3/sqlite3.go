package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Marshal results from SQLite3 types to Go types, then output JSON to the ResponseWriter.
func MarshalResults(rows *sql.Rows, w http.ResponseWriter) error {
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
