package common

import (
	"github.com/royallthefourth/bartlett"
	"net/http"
	"strings"
	"testing"
)

func dummyUserProvider(_ *http.Request) (interface{}, error) {
	return 1, nil
}

func TestSelect(t *testing.T) {
	table := bartlett.Table{
			Name:   `students_user`,
			UserID: `student_id`,
	}
	req, err := http.NewRequest(`GET`, `https://example.com/students_user`, strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}

	builder, err := Select(table, dummyUserProvider, req)
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
