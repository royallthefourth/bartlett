package bartlett

import (
	"database/sql"
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

	builder, err := b.select_(table, req)
	if err != nil {
		t.Fatal(err)
	}

	sql, args, err := builder.ToSql()
	if args[0] != 1 {
		t.Fatalf(`Expected userID arg to be 1 but got %v instead`, args[0])
	}
	if !strings.Contains(sql, `student_id =`) {
		t.Fatalf(`Expected query to require student_id but criterion not found in %s`, sql)
	}
}

//func TestAddColumns(t *testing.T) {
//	query := sqrl.Select(``)
//	query = addColumns(query, `age,score:grade`)
//	query.From(`students`)
//	sql, _, _ := query.ToSql()
//	t.Logf(sql)
//}
