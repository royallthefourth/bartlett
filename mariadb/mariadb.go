// Package mariadb provides a Bartlett driver for MariaDB databases.
package mariadb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/royallthefourth/bartlett"
	"net/http"
	"reflect"
)

// MariaDB provides logic specific to MariaDB and probably other MySQL compatibles, but MariaDB is the target.
type MariaDB struct {
	tables map[string][]column
}

type column struct {
	dataType reflect.Type
	name     string
}

type sqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default interface{}
	Extra   string
}

// GetColumns invokes `SHOW COLUMNS` and uses the output to determine valid columns for each table.
func (driver *MariaDB) GetColumns(db *sql.DB, t bartlett.Table) ([]string, error) {
	if driver.tables == nil {
		driver.tables = make(map[string][]column)
	}
	rows, err := db.Query(fmt.Sprintf(`SHOW COLUMNS FROM %s`, t.Name))
	if err != nil {
		return []string{}, err
	}

	columns := make([]string, 0)

	for rows.Next() {
		var c sqlColumn
		err = rows.Scan(&c.Field, &c.Type, &c.Null, &c.Key, &c.Default, &c.Extra)
		if err != nil {
			return columns, err
		}
		columns = append(columns, c.Field)
		driver.tables[t.Name] = append(driver.tables[t.Name], column{name: c.Field, dataType: mysqlTypeToGo(c.Type)})
	}

	return []string{}, err
}

// MarshalResults converts from MariaDB types to Go types, then outputs JSON to the ResponseWriter.
func (MariaDB) MarshalResults(rows *sql.Rows, w http.ResponseWriter) error {
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

// Driver gives weird results for column types by default.
// Let's pick our own types instead from https://github.com/go-sql-driver/mysql/blob/c45f530f8e7fe40f4687eaa50d0c8c5f1b66f9e0/fields.go#L16
func mysqlTypeToGo(t string) reflect.Type {
	switch t { // sure would like to have every branch of this converted to json by tests
	case `BIT`:
		return reflect.TypeOf(true)
	case `BLOB`:
		return reflect.TypeOf([]byte{})
	case `TEXT`:
		return reflect.TypeOf([]byte{})
	case `DATE`:
		return reflect.TypeOf(``)
	case `DATETIME`:
		return reflect.TypeOf(``)
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
		return reflect.TypeOf(``)
	case `TIMESTAMP`:
		return reflect.TypeOf(``)
	case `SMALLINT`:
		return reflect.TypeOf(int16(0))
	case `SET`:
		return reflect.TypeOf([]byte{})
	case `TINYINT`:
		return reflect.TypeOf(int8(0))
	case `YEAR`:
		return reflect.TypeOf(int16(0))
	default:
		return reflect.TypeOf(``)
	}
}
