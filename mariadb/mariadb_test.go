package mariadb

import (
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/royallthefourth/bartlett"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var dsn string

func init() {
	flag.StringVar(&dsn, `dsn`, ``, `MariaDB connection string`)
	flag.Parse()
}

func TestMariaDB(t *testing.T) {
	db, err := sql.Open(`mysql`, dsn)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(`CREATE TABLE students(student_id INTEGER PRIMARY KEY AUTO_INCREMENT, age INTEGER NOT NULL, grade INTEGER);`)
	if err != nil {
		t.Error(err)
	}
	defer db.Exec(`DROP TABLE students;`)

	_, err = db.Exec(`INSERT INTO students(age, grade) VALUES(18, 85),(20,91);`)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(`CREATE TABLE teachers(teacher_id INTEGER PRIMARY KEY AUTO_INCREMENT, name VARCHAR(25));`)
	if err != nil {
		t.Error(err)
	}
	defer db.Exec(`DROP TABLE teachers;`)

	_, err = db.Exec(`INSERT INTO teachers(name) VALUES('Mr. Smith'),('Ms. Key');`)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("CREATE TABLE todo(todo_id INTEGER PRIMARY KEY AUTO_INCREMENT, txt TEXT NOT NULL DEFAULT '');")
	if err != nil {
		t.Error(err)
	}
	defer db.Exec(`DROP TABLE todo;`)

	tables := []bartlett.Table{
		{
			Name:   `students`,
			UserID: `student_id`,
		},
		{
			Name: `teachers`,
		},
		{
			Name:     `todo`,
			Writable: true,
		},
	}

	b := bartlett.Bartlett{db, &MariaDB{}, tables, dummyUserProvider}

	routes := b.Routes()
	testSimpleGetAll(t, routes)
	testInsert(t, routes)
}

func dummyUserProvider(_ *http.Request) (interface{}, error) {
	return 1, nil
}

type student struct {
	Age       int `json:"age"`
	Grade     int `json:"grade"`
	StudentID int `json:"student_id"`
}

func testInsert(t *testing.T, routes []bartlett.Route) {
	req, err := http.NewRequest(`POST`, `https://example.com/todo`, strings.NewReader(`[{"txt":"hello"}]`))
	if err != nil {
		t.Error(err)
	}
	resp := httptest.NewRecorder()
	if routes[2].Path != `/todo` {
		t.Errorf(`Expected "todo" but got %s for path`, routes[2].Path)
	}

	routes[2].Handler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code`, resp.Code)
		t.Logf(resp.Body.String())
	}

	if !json.Valid(resp.Body.Bytes()) {
		t.Errorf(`Expected valid JSON response but got %s`, resp.Body.String())
	}

	req, err = http.NewRequest(`GET`, `https://example.com/todo`, strings.NewReader(``))
	if err != nil {
		t.Error(err)
	}
	resp = httptest.NewRecorder()
	routes[2].Handler(resp, req)

	if !strings.Contains(resp.Body.String(), `hello`) {
		t.Errorf(`Expected "hello" in response body but got %s`, resp.Body.String())
	}

	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code`, resp.Code)
	}

	if !json.Valid(resp.Body.Bytes()) {
		t.Errorf(`Expected valid JSON response but got %s`, resp.Body.String())
	}
}

func testSimpleGetAll(t *testing.T, routes []bartlett.Route) {
	req, err := http.NewRequest(`GET`, `https://example.com/students`, strings.NewReader(``))
	if err != nil {
		t.Error(err)
	}
	resp := httptest.NewRecorder()
	if routes[0].Path != `/students` {
		t.Errorf(`Expected "students" but got %s for path`, routes[0].Path)
	}

	routes[0].Handler(resp, req) // Fill the response

	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code`, resp.Code)
	}

	if !json.Valid(resp.Body.Bytes()) {
		t.Errorf(`Expected valid JSON response but got %s`, resp.Body.String())
	}

	testStudents := make([]student, 0)
	err = json.Unmarshal(resp.Body.Bytes(), &testStudents)
	if err != nil {
		t.Logf(resp.Body.String())
		t.Error(err)
	}

	if testStudents[0].Age != 18 {
		t.Errorf(`Expected first student to have age 18 but got %d instead`, testStudents[0].Age)
	}
}
