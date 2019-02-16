package bartlett

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	table := Table{
		Name:    `students_user`,
		UserID:  `student_id`,
		columns: []string{`grade`},
	}
	req, err := http.NewRequest(
		http.MethodPatch,
		`https://example.com/students_user?grade=not.eq.25`,
		strings.NewReader(`{"grade":25}`))
	if err != nil {
		t.Fatal(err)
	}

	b := Bartlett{&sql.DB{}, dummyDriver{}, []Table{table}, dummyUserProvider}

	builder, err := b.buildUpdate(table, req, 1, []byte(`{"grade":25}`))
	if err != nil {
		t.Fatal(err)
	}

	rawSQL, args, _ := builder.ToSql()
	if string(args[0].([]uint8)) != `25` {
		t.Errorf(`Expected userID arg to be 1 but got %+v instead`, args)
	}
	if !strings.Contains(rawSQL, `student_id =`) {
		t.Errorf(`Expected query to require student_id but criterion not found in %s`, rawSQL)
	}
}

func TestUpdateNeedsWhere(t *testing.T) {
	table := Table{
		Name:   `students_user`,
		UserID: `student_id`,
	}
	req, err := http.NewRequest(http.MethodPatch, `https://example.com/students_user`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}

	b := Bartlett{&sql.DB{}, dummyDriver{}, []Table{table}, dummyUserProvider}

	_, err = b.buildUpdate(table, req, 1, []byte{})
	if err == nil {
		t.Error(`Expected to error due to lack of constraints but got nil error`)
	}
}
