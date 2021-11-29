package expr

import (
	"fmt"
	"strings"
	"time"
)

// parseDescriptor returns a predefined schedule for the expression, or error if none matches.
func parseDescriptor(descriptor string, loc *time.Location) (Schedule, error) {
	switch descriptor {
	case "@yearly", "@annually":
		return &specSchedule{
			Second:   1 << secondBonds.min,
			Minute:   1 << minuteBounds.min,
			Hour:     1 << hourBounds.min,
			Dom:      1 << domBounds.min,
			Month:    1 << monthBounds.min,
			Dow:      allBits(dowBounds),
			Location: loc,
		}, nil

	case "@monthly":
		return &specSchedule{
			Second:   1 << secondBonds.min,
			Minute:   1 << minuteBounds.min,
			Hour:     1 << hourBounds.min,
			Dom:      1 << domBounds.min,
			Month:    allBits(monthBounds),
			Dow:      allBits(dowBounds),
			Location: loc,
		}, nil

	case "@weekly":
		return &specSchedule{
			Second:   1 << secondBonds.min,
			Minute:   1 << minuteBounds.min,
			Hour:     1 << hourBounds.min,
			Dom:      allBits(domBounds),
			Month:    allBits(monthBounds),
			Dow:      1 << dowBounds.min,
			Location: loc,
		}, nil

	case "@daily", "@midnight":
		return &specSchedule{
			Second:   1 << secondBonds.min,
			Minute:   1 << minuteBounds.min,
			Hour:     1 << hourBounds.min,
			Dom:      allBits(domBounds),
			Month:    allBits(monthBounds),
			Dow:      allBits(dowBounds),
			Location: loc,
		}, nil

	case "@hourly":
		return &specSchedule{
			Second:   1 << secondBonds.min,
			Minute:   1 << minuteBounds.min,
			Hour:     allBits(hourBounds),
			Dom:      allBits(domBounds),
			Month:    allBits(monthBounds),
			Dow:      allBits(dowBounds),
			Location: loc,
		}, nil
	}

	const every = "@every "
	if strings.HasPrefix(descriptor, every) {
		duration, err := time.ParseDuration(descriptor[len(every):])
		if err != nil {
			return nil, fmt.Errorf("expr: failed to parse duration %s: %s", descriptor, err)
		}
		return Every(duration), nil
	}

	return nil, fmt.Errorf("expr: unrecognized descriptor: %s", descriptor)
}
