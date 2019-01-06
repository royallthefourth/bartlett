// +build sqlite3

package sqlite3

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/royallthefourth/bartlett"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSQLite3(t *testing.T) {
	db, err := sql.Open(`sqlite3`, "file:test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE students(student_id INTEGER PRIMARY KEY AUTOINCREMENT, age INTEGER, grade INTEGER);`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO students(age, grade) VALUES(18, 85),(20,91);`)
	if err != nil {
		t.Fatal(err)
	}

	students := bartlett.Table{
		Name: `students`,
	}

	b := New(db, []bartlett.Table{students}, dummyUserProvider)

	testSimpleGetAll(t, b)
	testInvalidRequestMethod(t, b)
}

func dummyUserProvider(_ *http.Request) (interface{}, error) {
	return 0, nil
}

type student struct {
	Age       int `json:"age"`
	Grade     int `json:"grade"`
	StudentID int `json:"student_id"`
}

func testSimpleGetAll(t *testing.T, b bartlett.Bartlett) {
	paths, handlers := b.Routes()
	req, err := http.NewRequest(`GET`, `https://example.com/students`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	if paths[0] != `/students` {
		t.Fatalf(`Expected "students" but got %s for path`, paths[0])
	}

	handlers[0](resp, req) // Fill the response

	if resp.Code != http.StatusOK {
		t.Fatalf(`Expected "200" but got %d for status code`, resp.Code)
	}

	if !json.Valid(resp.Body.Bytes()) {
		t.Fatalf(`Expected valid JSON response but got %s`, resp.Body.String())
	}

	testStudents := make([]student, 0)
	err = json.Unmarshal(resp.Body.Bytes(), &testStudents)
	if err != nil {
		t.Logf(resp.Body.String())
		t.Fatal(err)
	}

	if testStudents[0].Age != 18 {
		t.Fatalf(`Expected first student to have age 18 but got %d instead`, testStudents[0].Age)
	}
}

func testInvalidRequestMethod(t *testing.T, b bartlett.Bartlett) {
	_, handlers := b.Routes()
	req, err := http.NewRequest(`POST`, `https://example.com/students`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()

	handlers[0](resp, req) // Fill the response

	if resp.Code != http.StatusNotImplemented {
		t.Fatalf(`Expected "501" but got %d for status code`, resp.Code)
	}
}
