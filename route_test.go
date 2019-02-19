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
	req, err := http.NewRequest(http.MethodGet, `https://example.com/teachers`, strings.NewReader(``))
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

func TestDeleteRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{
				Name:     `students`,
				Writable: true,
			},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	req, err := http.NewRequest(http.MethodDelete, `https://example.com/students`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf(`Expected "400" but got %d for status code`, resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPatchRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b := Bartlett{
		DB:     db,
		Driver: dummyDriver{},
		Tables: []Table{
			{
				Name:     `students`,
				Writable: true,
			},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	mock.ExpectExec(`UPDATE students SET name = \? WHERE id = \?`).
		WithArgs([]uint8(`todd`), `15`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req, err := http.NewRequest(http.MethodPatch, `https://example.com/students?id=eq.15`, strings.NewReader(`{"name":"todd"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code with body %s`, resp.Code, resp.Body.String())
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
			{
				Name:     `letters`,
				Writable: true,
			},
		},
		Users: dummyUserProvider,
	}

	_, handlers := b.Routes()

	mock.MatchExpectationsInOrder(false)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO letters`).
		WithArgs(`hello`, `5723`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	req, err := http.NewRequest(
		http.MethodPost,
		`https://example.com/letters`,
		strings.NewReader(`[{"a": "hello", "b": 5723}]`))
	if err != nil {
		t.Error(err)
	}
	resp := httptest.NewRecorder()
	handlers[0](resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf(`Expected "200" but got %d for status code in %s`, resp.Code, resp.Body.String())
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
		http.MethodPost,
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
		http.MethodPost,
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
		http.MethodPost,
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

func TestValidatePatch(t *testing.T) {
	b := Bartlett{}
	tbl := Table{Writable: true}
	req := http.Request{Method: http.MethodPatch}
	body := []byte(`[{"a":1}]`)
	status, _, _ := b.validateWrite(tbl, &req, body)
	if status != http.StatusBadRequest {
		t.Errorf(`Expected "400" but got %d for status code`, status)
	}
}

func TestValidatePost(t *testing.T) {
	b := Bartlett{}
	tbl := Table{Writable: true}
	req := http.Request{Method: http.MethodPost}
	body := []byte(`{"a":1}`)
	status, _, _ := b.validateWrite(tbl, &req, body)
	if status != http.StatusBadRequest {
		t.Errorf(`Expected "400" but got %d for status code`, status)
	}
}
