package bartlett

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{Name: `students`},
		},
		Users: dummyUserProvider,
	}

	routes, handlers := b.Routes()
	if routes[0] != `/students` {
		t.Errorf(`Expected route "students" but got %s instead`, routes[0])
	}

	mock.ExpectQuery(`SELECT \* FROM students`).WillReturnRows(sqlmock.NewRows([]string{`name`, `age`}))
	req, err := http.NewRequest(`GET`, `https://example.com/teachers`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf(`Expected "200" but got %d for status code`, resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{columns: []string{`a`, `b`},
				Name:     `letters`,
				Writable: true},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO letters`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	req, err := http.NewRequest(
		`POST`,
		`https://example.com/letters`,
		strings.NewReader(`[{"a": "hello", "b": 5723}]`))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code`, resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostReadOnly(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{Name: `letters`},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	req, err := http.NewRequest(
		`POST`,
		`https://example.com/letters`,
		strings.NewReader(`[{"a": "hello", "b": 5723}]`))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf(`Expected "405" but got %d for status code`, resp.Code)
	}
}

func TestPostInvalid(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{Name: `letters`, Writable: true},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	req, err := http.NewRequest(
		`POST`,
		`https://example.com/letters`,
		strings.NewReader(`[{"a": "hel`))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf(`Expected "400" but got %d for status code`, resp.Code)
	}
}

func TestPostForbidden(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{Name: `letters`, UserID: `user_id`, Writable: true},
		},
		Users: func(_ *http.Request) (interface{}, error) {
			return ``, errors.New(`invalid user`)
		},
	}

	_, handlers := b.Routes()
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO letters`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	req, err := http.NewRequest(
		`POST`,
		`https://example.com/letters`,
		strings.NewReader(`[{"a": "hello", "b": 5723}]`))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusForbidden {
		t.Errorf(`Expected "403" but got %d for status code`, resp.Code)
	}
}
