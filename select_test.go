package bartlett

import (
	"database/sql"
	"github.com/elgris/sqrl"
	"net/http"
	"strings"
	"testing"
)

func dummyUserProvider(_ *http.Request) (interface{}, error) {
	return 1, nil
}

type dummyDriver struct{}

func (d dummyDriver) MarshalResults(_ *sql.Rows, _ http.ResponseWriter) error {
	return nil
}

func (d dummyDriver) GetColumns(*sql.DB, Table) ([]string, error) {
	return []string{}, nil
}

func TestSelect(t *testing.T) {
	table := Table{
		Name:   `students_user`,
		UserID: `student_id`,
	}
	req, err := http.NewRequest(`GET`, `https://example.com/students_user`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}

	b := Bartlett{&sql.DB{}, dummyDriver{}, []Table{table}, dummyUserProvider}

	builder, err := b.buildSelect(table, req)
	if err != nil {
		t.Fatal(err)
	}

	rawSQL, args, _ := builder.ToSql()
	if args[0] != 1 {
		t.Fatalf(`Expected userID arg to be 1 but got %v instead`, args[0])
	}
	if !strings.Contains(rawSQL, `student_id =`) {
		t.Fatalf(`Expected query to require student_id but criterion not found in %s`, rawSQL)
	}
}

func TestParseColumns(t *testing.T) {
	schema := Table{columns: []string{`students`, `teachers`}}
	req, _ := http.NewRequest("GET", "http://example.com?select=students,parents", nil)

	cols := parseColumns(schema, req)
	if len(cols) != 1 {
		t.Errorf(`Expected 1 column in result but got %d`, len(cols))
	}

	if cols[0] != `students` {
		t.Errorf(`Expected table name "students" but got "%s"`, cols[0])
	}
}

func TestSelectOrder(t *testing.T) {
	schema := Table{columns: []string{`student_id`, `grade`}}
	req, _ := http.NewRequest("GET", "http://example.com?order=grade.asc,student_id", nil)
	query := sqrl.Select(`*`).From(`students`)
	query = selectOrder(query, schema, req)
	rawSQL, _, _ := query.ToSql()
	if !strings.Contains(rawSQL, `ORDER BY grade ASC, student_id DESC`) {
		t.Fatalf(`Expected "ORDER BY grade ASC, student_id DESC" but got %s`, rawSQL)
	}
}

func TestSelectLimit(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com?limit=10", nil)
	query := sqrl.Select(`*`).From(`students`)
	query = selectLimit(query, req)
	rawSQL, _, _ := query.ToSql()
	if !strings.Contains(rawSQL, `LIMIT 10 OFFSET 0`) {
		t.Errorf(`Expected "LIMIT 10 OFFSET 0" but got %s`, rawSQL)
	}

	req, _ = http.NewRequest("GET", "http://example.com?limit=10&offset=5", nil)
	query = sqrl.Select(`*`).From(`students`)
	query = selectLimit(query, req)
	rawSQL, _, _ = query.ToSql()
	if !strings.Contains(rawSQL, `LIMIT 10 OFFSET 5`) {
		t.Errorf(`Expected "LIMIT 10 OFFSET 0" but got %s`, rawSQL)
	}
}
