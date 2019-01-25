package bartlett

import (
	"testing"
)

func TestParseSimpleWhereCond(t *testing.T) {
	cond, val := parseSimpleWhereCond(`eq.90`)
	if cond != `eq` || val != `90` {
		t.Errorf(`Expected eq, 90 but got %s, %s`, cond, val)
	}

	cond, val = parseSimpleWhereCond(`not.eq.90`)
	if cond != `not.eq` || val != `90` {
		t.Errorf(`Expected not.eq, 90 but got %s, %s`, cond, val)
	}

	cond, val = parseSimpleWhereCond(`not.eq.hello,how.are.you`)
	if cond != `not.eq` || val != `hello,how.are.you` {
		t.Errorf(`Expected not.eq, hello,how.are.you but got %s, %s`, cond, val)
	}
}

func TestWhereIn(t *testing.T) {
	vals := whereIn(`10,20,30`)
	if vals[0] != `10` {
		t.Errorf(`Expected [10 20 30] but got %+v`, vals)
	}
}
