package bartlett

import (
	"database/sql"
	sqrl "github.com/Masterminds/squirrel"
	"net/http"
	"strings"
	"testing"
)

func TestDelete(t *testing.T) {
	table := Table{
		Name:    `students_user`,
		UserID:  `student_id`,
		columns: []string{`grade`},
	}
	req, err := http.NewRequest(http.MethodDelete, `https://example.com/students_user?grade=not.eq.25`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}

	b := Bartlett{&sql.DB{}, dummyDriver{}, []Table{table}, dummyUserProvider}

	builder, err := b.buildDelete(table, req)
	if err != nil {
		t.Fatal(err)
	}

	rawSQL, args, _ := builder.ToSql()
	if args[1] != 1 {
		t.Errorf(`Expected userID arg to be 1 but got %v instead`, args[0])
	}
	if !strings.Contains(rawSQL, `student_id =`) {
		t.Errorf(`Expected query to require student_id but criterion not found in %s`, rawSQL)
	}
}

func TestDeleteNeedsWhere(t *testing.T) {
	table := Table{
		Name:   `students_user`,
		UserID: `student_id`,
	}
	req, err := http.NewRequest(http.MethodDelete, `https://example.com/students_user`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}

	b := Bartlett{&sql.DB{}, dummyDriver{}, []Table{table}, dummyUserProvider}

	_, err = b.buildDelete(table, req)
	if err == nil {
		t.Error(`Expected to error due to lack of constraints but got nil error`)
	}
}

func TestDeleteLimit(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "http://example.com?limit=10", nil)
	query := sqrl.Delete(`*`).From(`students`)
	query = deleteLimit(query, req)
	rawSQL, _, _ := query.ToSql()
	if !strings.Contains(rawSQL, `LIMIT 10`) {
		t.Errorf(`Expected "LIMIT 10" but got %s`, rawSQL)
	}

	req, _ = http.NewRequest(http.MethodDelete, "http://example.com?limit=-1", nil)
	query = sqrl.Delete(`*`).From(`students`)
	query = deleteLimit(query, req)
	rawSQL, _, _ = query.ToSql()
	if strings.Contains(rawSQL, `LIMIT`) {
		t.Errorf(`Expected no LIMIT but got %s`, rawSQL)
	}

	req, _ = http.NewRequest(http.MethodDelete, "http://example.com?limit=asdf", nil)
	query = sqrl.Delete(`*`).From(`students`)
	query = deleteLimit(query, req)
	rawSQL, _, _ = query.ToSql()
	if strings.Contains(rawSQL, `LIMIT`) {
		t.Errorf(`Expected no LIMIT but got %s`, rawSQL)
	}
}

func TestDeleteOrder(t *testing.T) {
	schema := Table{columns: []string{`student_id`, `grade`}}
	req, _ := http.NewRequest(http.MethodDelete, "http://example.com?order=grade.asc,student_id", nil)
	query := sqrl.Delete(`*`).From(`students`)
	query = deleteOrder(query, schema, req)
	rawSQL, _, _ := query.ToSql()
	if !strings.Contains(rawSQL, `ORDER BY grade ASC, student_id DESC`) {
		t.Fatalf(`Expected "ORDER BY grade ASC, student_id DESC" but got %s`, rawSQL)
	}
}

func TestDeleteWhere(t *testing.T) {
	schema := Table{Name: `students`, columns: []string{`student_id`, `grade`}}
	req, _ := http.NewRequest(
		http.MethodDelete,
		"http://example.com/students?grade=eq.90&student_id=not.eq.25&student_id=in.(10,20,30)&student_id=not.in.(11,12)&grade=like.a*c",
		nil)
	query, _ := deleteWhere(sqrl.Delete(`students`), schema, req)
	rawSQL, _, _ := query.ToSql()
	if !strings.Contains(rawSQL, `student_id != ?`) || !strings.Contains(rawSQL, `grade = ?`) {
		t.Errorf(`Expected "grade = ? AND student_id != ?" but got %s`, rawSQL)
	}
	if !strings.Contains(rawSQL, `IN (?,?,?)`) {
		t.Errorf(`Expected "IN (?,?,?)" but got %s`, rawSQL)
	}
	if !strings.Contains(rawSQL, `NOT IN (?,?)`) {
		t.Errorf(`Expected "NOT IN (?,?)" but got %s`, rawSQL)
	}
	if !strings.Contains(rawSQL, `LIKE ?`) {
		t.Errorf(`Expected "IN ?" but got %s`, rawSQL)
	}
}
