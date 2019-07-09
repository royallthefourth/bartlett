package bartlett

import (
	"database/sql"
	"testing"
)

func TestBartlett_ProbeTables(t *testing.T) {
	tables := []Table{
		{
			Name:     `students`,
			Writable: true,
		},
	}

	b := Bartlett{DB: &sql.DB{}, Driver: dummyDriver{}, Tables: tables, Users: dummyUserProvider}
	b.ProbeTables(false)
	if b.Tables[0].Name != `students` || b.Tables[0].Writable != true {
		t.Errorf(`Expected students to be writable but got %s as %t`, b.Tables[0].Name, b.Tables[0].Writable)
	}
	if b.Tables[1].Name != `teachers` || b.Tables[1].Writable != false {
		t.Errorf(`Expected teachers to be non-writable but got %s as %t`, b.Tables[0].Name, b.Tables[0].Writable)
	}
}
