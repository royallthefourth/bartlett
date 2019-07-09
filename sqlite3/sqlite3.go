// Package sqlite3 provides a Bartlett driver for SQLite3 databases.
package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/royallthefourth/bartlett"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

// SQLite3 provides logic specific to SQLite3 databases.
type SQLite3 struct {
	tables map[string][]column
}

type column struct {
	dataType reflect.Type
	name     string
}

// GetColumns queries `sqlite_master` and returns a list of valid column names.
func (driver *SQLite3) GetColumns(db *sql.DB, t bartlett.Table) ([]string, error) {
	if driver.tables == nil {
		driver.tables = make(map[string][]column)
	}
	var (
		createQuery string
		out         []string
	)
	rows, err := sqrl.Select(`sql`).From(`sqlite_master`).Where(`name = ?`, t.Name).RunWith(db).Query()
	if err != nil {
		return []string{}, err
	}

	rows.Next() // We should only expect a single row here.
	err = rows.Scan(&createQuery)
	if err != nil {
		return []string{}, err
	}

	driver.tables[t.Name] = parseCreateTable(createQuery)
	for _, col := range driver.tables[t.Name] {
		out = append(out, col.name)
	}

	return out, err
}

// MarshalResults converts results from SQLite3 types to Go types, then outputs JSON to the ResponseWriter.
func (driver SQLite3) MarshalResults(rows *sql.Rows, w http.ResponseWriter) error {
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
			types[i] = dbTypeToGoType(columnType.DatabaseTypeName())
		}
	}

	values := make([]interface{}, len(columnTypes))
	data := make(map[string]interface{})

	_, err = w.Write([]byte{'['}) // Start the output array
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

func (driver *SQLite3) ProbeTables(db *sql.DB) []bartlett.Table {
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	tables := make([]bartlett.Table, 0)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}

		tables = append(tables, bartlett.Table{Name: name})
	}

	return tables
}

func dbTypeToGoType(dbType string) reflect.Type {
	t := strings.TrimSpace(strings.ToLower(dbType))
	if strings.Contains(t, `int`) {
		if strings.Contains(t, `unsigned`) {
			return reflect.TypeOf(uint(0))
		}
		return reflect.TypeOf(int(0))
	} else if strings.Contains(t, `char`) {
		return reflect.TypeOf(``)
	} else if strings.Contains(t, `real`) ||
		strings.Contains(t, `double`) ||
		strings.Contains(t, `float`) {
		return reflect.TypeOf(float64(0))
	}
	return reflect.TypeOf([]byte{}) // Guess it's a blob
}

func parseCreateTable(sql string) (columns []column) {
	colSpec := regexp.MustCompile(`.*CREATE\s+TABLE\s+(\S+)\s*\((.*)\).*`)
	firstWord := regexp.MustCompile(`\s.*`)
	specs := strings.Split(colSpec.FindStringSubmatch(sql)[2], `,`)
	for _, spec := range specs {
		colName := firstWord.ReplaceAllString(strings.TrimSpace(spec), ``)
		rawType := firstWord.FindString(spec)
		columns = append(columns, column{
			dataType: dbTypeToGoType(rawType),
			name:     colName,
		})
	}

	return columns
}
