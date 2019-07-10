package bartlett

import (
	"reflect"
	"strings"
	"testing"
)

func TestPrepareInsert(t *testing.T) {
	tbl := Table{
		columns:  []string{`a`, `b`},
		Name:     `letters`,
		Writable: true,
	}
	sql, args, err := tbl.prepareInsert([]byte(`{"a": "test", "b": 5723, "c": "disregard"}`), 1, nil).ToSql()
	if err != nil {
		t.Errorf(err.Error())
	}

	if !strings.Contains(sql, `INSERT INTO letters (a,b) VALUES (?,?)`) {
		t.Errorf(`Expected "INSERT INTO letters (a,b) VALUES (?,?)" but got %s`, sql)
	}

	if args[0] != `test` {
		t.Errorf(`Expected "test" but got %s`, args[0])
	}
}

func TestPrepareInsertUserID(t *testing.T) {
	tbl := Table{
		columns:  []string{`a`, `b`, `userID`},
		Name:     `letters`,
		Writable: true,
		UserID:   `userID`,
	}
	sql, args, err := tbl.prepareInsert([]byte(`{"a": "test", "b": 5723, "userID": "disregard"}`), 1, nil).ToSql()
	if err != nil {
		t.Errorf(err.Error())
	}

	if !strings.Contains(sql, `INSERT INTO letters (a,b,userID) VALUES (?,?,?)`) {
		t.Errorf(`Expected "INSERT INTO letters (a,b,userID) VALUES (?,?,?)" but got %s`, sql)
	}

	if !reflect.DeepEqual(args[0], `test`) {
		t.Errorf(`Expected "test" but got %s`, args[0])
	}

	if !reflect.DeepEqual(args[1], `5723`) {
		t.Errorf(`Expected 5723 but got %c`, args[1])
	}

	if !reflect.DeepEqual(args[2], 1) {
		t.Errorf(`Expected 1 but got %c`, args[2])
	}
}

func TestPrepareInsertIDColumn(t *testing.T) {
	tbl := Table{
		columns: []string{`a`, `b`, `letter_id`},
		IDColumn: IDSpec{
			Name:      `letter_id`,
			Generator: func() interface{} { return 1 },
		},
		Name:     `letters`,
		Writable: true,
	}
	sql, args, err := tbl.prepareInsert([]byte(`{"a": "test", "b": 5723, "letter_id": "disregard"}`), 1, 1).ToSql()
	if err != nil {
		t.Errorf(err.Error())
	}

	if !strings.Contains(sql, `INSERT INTO letters (a,b,letter_id) VALUES (?,?,?)`) {
		t.Errorf(`Expected "INSERT INTO letters (a,b,letter_id) VALUES (?,?,?)" but got %s`, sql)
	}

	if !reflect.DeepEqual(args[2].(int), int(1)) {
		t.Errorf(`Expected 1 but got %c`, args[2])
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
