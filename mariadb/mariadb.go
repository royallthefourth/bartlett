package mariadb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/royallthefourth/bartlett"
	"github.com/royallthefourth/bartlett/common"
	"log"
	"net/http"
	"reflect"
	"time"
)

type MariaDB struct {
	db     *sql.DB
	tables []bartlett.Table
	users  bartlett.UserIDProvider
}

func New(db *sql.DB, tables []bartlett.Table, users bartlett.UserIDProvider) MariaDB {
	return MariaDB{
		db:     db,
		tables: tables,
		users:  users,
	}
}

func (b MariaDB) Routes() (paths []string, handlers []func(http.ResponseWriter, *http.Request)) {
	return common.Routes(b.db, b.users, handleRoute, b.tables)
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

// Driver gives weird results for column types by default.
// Let's pick our own types instead from https://github.com/go-sql-driver/mysql/blob/c45f530f8e7fe40f4687eaa50d0c8c5f1b66f9e0/fields.go#L16
func mysqlTypeToGo(t string) reflect.Type {
	switch t {	// sure would like to have every branch of this converted to json by tests
	case `BIT`:
		return reflect.TypeOf(true)
	case `BLOB`:
		return reflect.TypeOf([]byte{})
	case `TEXT`:
		return reflect.TypeOf([]byte{})
	case `DATE`:
		return reflect.TypeOf(time.Now())
	case `DATETIME`:
		return reflect.TypeOf(time.Now())
	case `DECIMAL`:
		return reflect.TypeOf(``)
	case `DOUBLE`:
		return reflect.TypeOf(float64(0))
	case `ENUM`:
		return reflect.TypeOf(``)
	case `FLOAT`:
		return reflect.TypeOf(float32(0))
	case `GEOMETRY`:
		return reflect.TypeOf([]byte{})
	case `MEDIUMINT`:
		return reflect.TypeOf(int32(0))
	case `JSON`:
		return reflect.TypeOf([]byte{})
	case `INT`:
		return reflect.TypeOf(int(0))
	case `LONGTEXT`:
		return reflect.TypeOf([]byte{})
	case `LONGBLOB`:
		return reflect.TypeOf([]byte{})
	case `BIGINT`:
		return reflect.TypeOf(int64(0))
	case `MEDIUMTEXT`:
		return reflect.TypeOf([]byte{})
	case `MEDIUMBLOB`:
		return reflect.TypeOf([]byte{})
	case `CHAR`:
		return reflect.TypeOf(``)
	case `BINARY`:
		return reflect.TypeOf([]byte{})
	case `VARCAHR`:
		return reflect.TypeOf(``)
	case `VARBINARY`:
		return reflect.TypeOf([]byte{})
	case `TIME`:
		return reflect.TypeOf(time.Now())
	case `TIMESTAMP`:
		return reflect.TypeOf(time.Now())
	case `SMALLINT`:
		return reflect.TypeOf(int16(0))
	case `SET`:
		return reflect.TypeOf([]byte{})
	case `TINYINT`:
		return reflect.TypeOf(int8(0))
	case `YEAR`:
		return reflect.TypeOf(int16(0))
	default:
		return nil
	}
}// TODO establish list of mysql driver types

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
		types[i] = mysqlTypeToGo(columnType.DatabaseTypeName())

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
