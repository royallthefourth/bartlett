package bartlett

import (
	"strings"
	"testing"
)

func TestPrepareInsert(t *testing.T) {
	tbl := Table{
		columns:  []string{`a`, `b`},
		Name:     `letters`,
		Writable: true,
	}
	sql, _, err := tbl.prepareInsert([]byte(`{"a": "test", "b": 5723, "c": "disregard"}`), 1).ToSql()
	if err != nil {
		t.Errorf(err.Error())
	}

	if !strings.Contains(sql, `INSERT INTO letters (a,b)`) {
		t.Errorf(`Expected "INSERT INTO letters (a,b)" but got %s`, sql)
	}
}

func TestValidWriteColumns(t *testing.T) {
	idTable := Table{
		columns:  []string{`a`, `b`, `c`},
		Name:     `id`,
		IDColumn: IDSpec{Name: `a`, Generator: func() interface{} { return 1 }},
		Writable: true,
		UserID:   `b`,
	}
	idCols := idTable.validWriteColumns()
	if idCols[0] != `c` || len(idCols) != 1 {
		t.Errorf(`Expected [c] but got %+v instead`, idCols)
	}

	nonIDTable := Table{
		columns:  []string{`a`, `b`, `c`},
		Name:     `non_id`,
		Writable: true,
	}
	nonIDCols := nonIDTable.validWriteColumns()
	if nonIDCols[0] != `a` || len(nonIDCols) != 3 {
		t.Errorf(`Expected [a, b, c] but got %+v instead`, idCols)
	}
}
