package expr

// bounds provides a range of acceptable values (plus a map of name to value).
type bounds struct {
	min   uint
	max   uint
	names map[string]uint
}

// The bounds for each field.
var (
	secondBonds  = bounds{0, 59, nil}
	minuteBounds = bounds{0, 59, nil}
	hourBounds   = bounds{0, 23, nil}
	domBounds    = bounds{1, 31, nil}
	monthBounds  = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dowBounds = bounds{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)
