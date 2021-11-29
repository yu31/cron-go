package expr

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Option Configuration options for creating a expr. Most options specify which
// fields should be included, while others enable features. If a field is not
// included the expr will assume a default value. These options do not change
// the order fields are parse in.
type Option int

const (
	Second     Option = 1 << iota // Seconds field, default 0
	Minute                        // Minutes field, default 0
	Hour                          // Hours field, default 0
	Dom                           // Day of month field, default *
	Month                         // Month field, default *
	Dow                           // Day of week field, default *
	Descriptor                    // Allow descriptors such as @monthly, @weekly, etc.
)

var places = []Option{
	Second,
	Minute,
	Hour,
	Dom,
	Month,
	Dow,
}

var defaults = []string{
	"0",
	"0",
	"0",
	"*",
	"*",
	"*",
}

// Parser A custom parser that can be configured.
type Parser struct {
	options Option
}

// New creates a Parser with custom options.
//
// It panics if more than one Optional is given, since it would be impossible to
// correctly infer which optional is provided or missing in general.
//
// Examples
//
//  // Standard expr without descriptors
//  specParser := New(Minute | Hour | Dom | Month | Dow)
//  sched, err := expr.Parse("0 0 15 */3 *")
//
//  // Same as above, just excludes time fields
//  subsParser := New(Dom | Month | Dow)
//  sched, err := expr.Parse("15 */3 *")
//
func New(options Option) Parser {
	return Parser{options: options}
}

// Parse returns a new crontab schedule representing the given spec.
// It returns a descriptive error if the spec is not valid.
// It accepts crontab specs and features configured by New.
func (p Parser) Parse(spec string) (Schedule, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("expr: empty spec string")
	}

	// Extract timezone if present
	var loc = time.Local
	if strings.HasPrefix(spec, "TZ=") || strings.HasPrefix(spec, "CRON_TZ=") {
		var err error
		i := strings.Index(spec, " ")
		eq := strings.Index(spec, "=")
		if loc, err = time.LoadLocation(spec[eq+1 : i]); err != nil {
			return nil, fmt.Errorf("expr: provided bad location %s: %v", spec[eq+1:i], err)
		}
		spec = strings.TrimSpace(spec[i:])
	}

	// Handle named schedules (descriptors), if configured
	if strings.HasPrefix(spec, "@") {
		if p.options&Descriptor == 0 {
			return nil, fmt.Errorf("expr: does not accept descriptors: %v", spec)
		}
		return parseDescriptor(spec, loc)
	}

	// Split on whitespace.
	fields := strings.Fields(spec)
	// Validate & fill in any omitted or optional fields
	var err error
	fields, err = normalizeFields(fields, p.options)
	if err != nil {
		return nil, err
	}

	field := func(field string, r bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(field, r)
		return bits
	}

	var (
		second = field(fields[0], secondBonds)
		minute = field(fields[1], minuteBounds)
		hour   = field(fields[2], hourBounds)
		dom    = field(fields[3], domBounds)
		month  = field(fields[4], monthBounds)
		dow    = field(fields[5], dowBounds)
	)
	if err != nil {
		return nil, err
	}

	return &specSchedule{
		Second:   second,
		Minute:   minute,
		Hour:     hour,
		Dom:      dom,
		Month:    month,
		Dow:      dow,
		Location: loc,
	}, nil
}

// normalizeFields takes a subset set of the time fields and returns the full set
// with defaults (zeroes) populated for unset fields.
//
// As part of performing this function, it also validates that the provided
// fields are compatible with the configured options.
func normalizeFields(fields []string, options Option) ([]string, error) {
	// Figure out how many fields we need
	num := 0
	for _, place := range places {
		if options&place > 0 {
			num++
		}
	}
	// Validate number of fields
	if len(fields) != num {
		return nil, fmt.Errorf("expr: expected exactly %d fields, found %d: %s", num, len(fields), fields)
	}

	// Populate all fields not part of options with their defaults
	n := 0
	expandedFields := make([]string, len(places))
	copy(expandedFields, defaults)
	for i, place := range places {
		if options&place > 0 {
			expandedFields[i] = fields[n]
			n++
		}
	}
	return expandedFields, nil
}

// getField returns an Int with the bits set representing all of the times that
// the field represents or error parsing field value.  A "field" is a comma-separated
// list of "ranges".
func getField(field string, r bounds) (uint64, error) {
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		bit, err := getRange(expr, r)
		if err != nil {
			return bits, err
		}
		bits |= bit
	}
	return bits, nil
}

// getRange returns the bits indicated by the given expression:
//   number | number "-" number [ "/" number ]
// or error parsing range.
func getRange(expr string, r bounds) (uint64, error) {
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
		err              error
	)

	var extra uint64
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra = starBit
	} else {
		start, err = parseIntOrName(lowAndHigh[0], r.names)
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end, err = parseIntOrName(lowAndHigh[1], r.names)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("expr: too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = mustParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}

		// Special handling: "N/step" means "N-max/step".
		if singleDigit {
			end = r.max
		}
		if step > 1 {
			extra = 0
		}
	default:
		return 0, fmt.Errorf("expr: too many slashes: %s", expr)
	}

	if start < r.min {
		return 0, fmt.Errorf("expr: beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		return 0, fmt.Errorf("expr: end of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		return 0, fmt.Errorf("expr: beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("expr: step of range should be a positive number: %s", expr)
	}

	return getBits(start, end, step) | extra, nil
}

// parseIntOrName returns the (possibly-named) integer contained in expr.
func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt, nil
		}
	}
	return mustParseInt(expr)
}

// mustParseInt parses the given expression as an int or returns an error.
func mustParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		return 0, fmt.Errorf("expr: failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("expr: negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num), nil
}

// getBits sets all bits in the range [min, max], modulo the given step size.
func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// allBits returns all bits within the given bounds.  (plus the star bit)
func allBits(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}
