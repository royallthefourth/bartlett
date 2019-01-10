package bartlett

import (
	"net/http"
	"testing"
)

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
