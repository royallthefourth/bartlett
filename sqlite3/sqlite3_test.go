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

	_, err = db.Exec(`CREATE TABLE students(student_id INTEGER PRIMARY KEY AUTOINCREMENT, age INTEGER NOT NULL, grade INTEGER);`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO students(age, grade) VALUES(18, 85),(20,91);`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE teachers(teacher_id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR);`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO teachers(name) VALUES('Mr. Smith'),('Ms. Key');`)
	if err != nil {
		t.Fatal(err)
	}

	tables := []bartlett.Table{
		{
			Name:     `students`,
			UserID:   `student_id`,
			Writable: true,
		},
		{
			Name:     `teachers`,
			Writable: true,
		},
	}

	b := bartlett.Bartlett{db, &SQLite3{}, tables, dummyUserProvider}

	testSimpleGetAll(t, b)
	testUserGetAll(t, b)
	testGetColumn(t, b)
}

func dummyUserProvider(_ *http.Request) (interface{}, error) {
	return 1, nil
}

type student struct {
	Age       int `json:"age"`
	Grade     int `json:"grade"`
	StudentID int `json:"student_id"`
}

type teacher struct {
	Name      string `json:"name"`
	TeacherID int    `json:"teacher_id"`
}

func testGetColumn(t *testing.T, b bartlett.Bartlett) {
	paths, handlers := b.Routes()
	req, err := http.NewRequest(`GET`, `https://example.com/teachers?select=teacher_id`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()

	for i, path := range paths {
		if path == `/teachers` {
			handlers[i](resp, req) // Fill the response

			if !json.Valid(resp.Body.Bytes()) {
				t.Fatalf(`Expected valid JSON response but got %s`, resp.Body.String())
			}

			if strings.Contains(resp.Body.String(), `name`) {
				t.Fatalf(`Expected only IDs but got %s instead`, resp.Body.String())
			}
		}
	}
}

func testSimpleGetAll(t *testing.T, b bartlett.Bartlett) {
	paths, handlers := b.Routes()
	req, err := http.NewRequest(`GET`, `https://example.com/teachers`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()

	for i, path := range paths {
		if path == `/teachers` {
			handlers[i](resp, req) // Fill the response

			if !json.Valid(resp.Body.Bytes()) {
				t.Errorf(`Expected valid JSON response but got %s`, resp.Body.String())
			}

			teachers := make([]teacher, 0)
			err = json.Unmarshal(resp.Body.Bytes(), &teachers)
			if err != nil {
				t.Logf(resp.Body.String())
				t.Errorf(err.Error())
			}

			if teachers[0].Name != `Mr. Smith` {
				t.Errorf(`Expected first student to have age 18 but got %s instead`, teachers[0].Name)
			}
		}
	}
}

func testUserGetAll(t *testing.T, b bartlett.Bartlett) {
	paths, handlers := b.Routes()
	req, err := http.NewRequest(`GET`, `https://example.com/students`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()

	for i, path := range paths {
		if path == `/students` {
			handlers[i](resp, req) // Fill the response

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

			if len(testStudents) != 1 {
				t.Fatalf(`Expected exactly 1 result but got %d instead`, len(testStudents))
			}

			if testStudents[0].Age != 18 {
				t.Fatalf(`Expected first student to have age 18 but got %d instead`, testStudents[0].Age)
			}

			if testStudents[0].StudentID != 1 {
				t.Fatalf(`Expected student to have student_id 1 but got %d instead`, testStudents[0].StudentID)
			}
		}
	}
}

func TestParseCreateTable(t *testing.T) {
	columns := parseCreateTable(`CREATE TABLE students(age int NOT NULL, grade INT)`)
	if columns[0].name != `age` || columns[1].name != `grade` {
		t.Errorf(`Expected "age" and "grade" but got %+v`, columns)
	}
}
