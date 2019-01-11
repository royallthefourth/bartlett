package bartlett

import (
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
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
}
