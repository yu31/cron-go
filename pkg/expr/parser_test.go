package expr

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var secondParser = New(Second | Minute | Hour | Dom | Month | Dow | Descriptor)

func TestExpr_Range(t *testing.T) {
	zero := uint64(0)
	ranges := []struct {
		expr     string
		min, max uint
		expected uint64
		err      string
	}{
		{"5", 0, 7, 1 << 5, ""},
		{"0", 0, 7, 1 << 0, ""},
		{"7", 0, 7, 1 << 7, ""},

		{"5-5", 0, 7, 1 << 5, ""},
		{"5-6", 0, 7, 1<<5 | 1<<6, ""},
		{"5-7", 0, 7, 1<<5 | 1<<6 | 1<<7, ""},

		{"5-6/2", 0, 7, 1 << 5, ""},
		{"5-7/2", 0, 7, 1<<5 | 1<<7, ""},
		{"5-7/1", 0, 7, 1<<5 | 1<<6 | 1<<7, ""},

		{"*", 1, 3, 1<<1 | 1<<2 | 1<<3 | starBit, ""},
		{"*/2", 1, 3, 1<<1 | 1<<3, ""},

		{"5--5", 0, 0, zero, "too many hyphens"},
		{"jan-x", 0, 0, zero, "failed to parse int from"},
		{"2-x", 1, 5, zero, "failed to parse int from"},
		{"*/-12", 0, 0, zero, "negative number"},
		{"*//2", 0, 0, zero, "too many slashes"},
		{"1", 3, 5, zero, "below minimum"},
		{"6", 3, 5, zero, "above maximum"},
		{"5-3", 3, 5, zero, "beyond end of range"},
		{"*/0", 0, 0, zero, "should be a positive number"},
	}

	for _, c := range ranges {
		actual, err := getRange(c.expr, bounds{c.min, c.max, nil})
		if len(c.err) != 0 && (err == nil || !strings.Contains(err.Error(), c.err)) {
			t.Errorf("%s => expected %v, got %v", c.expr, c.err, err)
		}
		if len(c.err) == 0 && err != nil {
			t.Errorf("%s => unexpected error %v", c.expr, err)
		}
		if actual != c.expected {
			t.Errorf("%s => expected %d, got %d", c.expr, c.expected, actual)
		}
	}
}

func TestExpr_Field(t *testing.T) {
	fields := []struct {
		expr     string
		min, max uint
		expected uint64
	}{
		{"5", 1, 7, 1 << 5},
		{"5,6", 1, 7, 1<<5 | 1<<6},
		{"5,6,7", 1, 7, 1<<5 | 1<<6 | 1<<7},
		{"1,5-7/2,3", 1, 7, 1<<1 | 1<<5 | 1<<7 | 1<<3},
	}

	for _, c := range fields {
		actual, _ := getField(c.expr, bounds{c.min, c.max, nil})
		if actual != c.expected {
			t.Errorf("%s => expected %d, got %d", c.expr, c.expected, actual)
		}
	}
}

func TestExpr_All(t *testing.T) {
	all := []struct {
		r        bounds
		expected uint64
	}{
		{minuteBounds, 0xfffffffffffffff}, // 0-59: 60 ones
		{hourBounds, 0xffffff},            // 0-23: 24 ones
		{domBounds, 0xfffffffe},           // 1-31: 31 ones, 1 zero
		{monthBounds, 0x1ffe},             // 1-12: 12 ones, 1 zero
		{dowBounds, 0x7f},                 // 0-6: 7 ones
	}

	for _, c := range all {
		actual := allBits(c.r) // allBits() adds the starBit, so compensate for that..
		if c.expected|starBit != actual {
			t.Errorf("%d-%d/%d => expected %b, got %b",
				c.r.min, c.r.max, 1, c.expected|starBit, actual)
		}
	}
}

func TestExpr_Bits(t *testing.T) {
	bits := []struct {
		min, max, step uint
		expected       uint64
	}{
		{0, 0, 1, 0x1},
		{1, 1, 1, 0x2},
		{1, 5, 2, 0x2a}, // 101010
		{1, 4, 2, 0xa},  // 1010
	}

	for _, c := range bits {
		actual := getBits(c.min, c.max, c.step)
		if c.expected != actual {
			t.Errorf("%d-%d/%d => expected %b, got %b",
				c.min, c.max, c.step, c.expected, actual)
		}
	}
}

func TestExpr_ParseScheduleErrors(t *testing.T) {
	var tests = []struct{ expr, err string }{
		{"* 5 j * * *", "failed to parse int from"},
		{"@every Xm", "failed to parse duration"},
		{"@unrecognized", "unrecognized descriptor"},
		{"", "empty spec string"},
	}
	for _, c := range tests {
		actual, err := secondParser.Parse(c.expr)
		if err == nil || !strings.Contains(err.Error(), c.err) {
			t.Errorf("%s => expected %v, got %v", c.expr, c.err, err)
		}
		if actual != nil {
			t.Errorf("expected nil schedule on error, got %v", actual)
		}
	}
}

func TestExpr_ParseSchedule(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	entries := []struct {
		parser   Parser
		expr     string
		expected Schedule
	}{
		{secondParser, "0 5 * * * *", every5min(time.Local)},
		{Standard, "5 * * * *", every5min(time.Local)},
		{secondParser, "TZ=UTC  0 5 * * * *", every5min(time.UTC)},
		{Standard, "TZ=UTC  5 * * * *", every5min(time.UTC)},
		{secondParser, "TZ=Asia/Tokyo 0 5 * * * *", every5min(tokyo)},
		{secondParser, "@every 5m", everySchedule{5 * time.Minute}},
		{secondParser, "@midnight", midnight(time.Local)},
		{secondParser, "TZ=UTC  @midnight", midnight(time.UTC)},
		{secondParser, "TZ=Asia/Tokyo @midnight", midnight(tokyo)},
		{secondParser, "@yearly", annual(time.Local)},
		{secondParser, "@annually", annual(time.Local)},
		{
			parser: secondParser,
			expr:   "* 5 * * * *",
			expected: &specSchedule{
				Second:   allBits(secondBonds),
				Minute:   1 << 5,
				Hour:     allBits(hourBounds),
				Dom:      allBits(domBounds),
				Month:    allBits(monthBounds),
				Dow:      allBits(dowBounds),
				Location: time.Local,
			},
		},
	}

	for _, c := range entries {
		actual, err := c.parser.Parse(c.expr)
		if err != nil {
			t.Errorf("%s => unexpected error %v", c.expr, err)
		}
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("%s => expected %b, got %b", c.expr, c.expected, actual)
		}
	}
}

func TestExpr_NormalizeFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		options  Option
		expected []string
	}{
		{
			"AllFields_NoOptional",
			[]string{"0", "5", "*", "*", "*", "*"},
			Second | Minute | Hour | Dom | Month | Dow | Descriptor,
			[]string{"0", "5", "*", "*", "*", "*"},
		},
		{
			"SubsetFields_NoOptional",
			[]string{"5", "15", "*"},
			Hour | Dom | Month,
			[]string{"0", "0", "5", "15", "*", "*"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := normalizeFields(test.input, test.options)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestExpr_NormalizeFields_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		options Option
		err     string
	}{
		{
			"TooManyFields",
			[]string{"0", "5", "*", "*"},
			Second | Minute | Hour,
			"",
		},
		{
			"NoFields",
			[]string{},
			Second | Minute | Hour,
			"",
		},
		{
			"TooFewFields",
			[]string{"*"},
			Second | Minute | Hour,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := normalizeFields(test.input, test.options)
			if err == nil {
				t.Errorf("expected an error, got none. results: %v", actual)
				t.FailNow()
			}
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("expected error %q, got %q", test.err, err.Error())
			}
		})
	}
}

func TestExpr_StandardSpecSchedule(t *testing.T) {
	entries := []struct {
		expr     string
		expected Schedule
		err      string
	}{
		{
			expr:     "5 * * * *",
			expected: &specSchedule{1 << secondBonds.min, 1 << 5, allBits(hourBounds), allBits(domBounds), allBits(monthBounds), allBits(dowBounds), time.Local},
		},
		{
			expr:     "@every 5m",
			expected: everySchedule{time.Duration(5) * time.Minute},
		},
		{
			expr: "5 j * * *",
			err:  "failed to parse int from",
		},
		{
			expr: "* * * *",
			err:  "expected exactly 5 fields",
		},
	}

	for _, c := range entries {
		actual, err := Standard.Parse(c.expr)
		if len(c.err) != 0 && (err == nil || !strings.Contains(err.Error(), c.err)) {
			t.Errorf("%s => expected %v, got %v", c.expr, c.err, err)
		}
		if len(c.err) == 0 && err != nil {
			t.Errorf("%s => unexpected error %v", c.expr, err)
		}
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("%s => expected %b, got %b", c.expr, c.expected, actual)
		}
	}
}

func TestExpr_NoDescriptorParser(t *testing.T) {
	parser := New(Minute | Hour)
	_, err := parser.Parse("@every 1m")
	if err == nil {
		t.Error("expected an error, got none")
	}
}

func every5min(loc *time.Location) *specSchedule {
	return &specSchedule{1 << 0, 1 << 5, allBits(hourBounds), allBits(domBounds), allBits(monthBounds), allBits(dowBounds), loc}
}

//func every5min5s(loc *time.Location) *specSchedule {
//	return &specSchedule{1 << 5, 1 << 5, allBits(hourBounds), allBits(domBounds), allBits(monthBounds), allBits(dowBounds), loc}
//}

func midnight(loc *time.Location) *specSchedule {
	return &specSchedule{1, 1, 1, allBits(domBounds), allBits(monthBounds), allBits(dowBounds), loc}
}

func annual(loc *time.Location) *specSchedule {
	return &specSchedule{
		Second:   1 << secondBonds.min,
		Minute:   1 << minuteBounds.min,
		Hour:     1 << hourBounds.min,
		Dom:      1 << domBounds.min,
		Month:    1 << monthBounds.min,
		Dow:      allBits(dowBounds),
		Location: loc,
	}
}
