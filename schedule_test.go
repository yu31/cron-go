package gcron

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSchedule_UnixCron1(t *testing.T) {
	var err error
	var begin time.Time
	var end time.Time

	begin, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:00:00")
	require.Nil(t, err)
	end, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
	require.Nil(t, err)

	begin = begin.In(time.Local)
	end = end.In(time.Local)

	sch := UnixCron{
		Begin:        begin,
		End:          end,
		Express:      "*/5 * * * *",
		once:         sync.Once{},
		exprSchedule: nil,
	}

	t.Run("TestNext1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:05:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), current.Add(time.Minute*5).String())
	})

	t.Run("TestBegin1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:00:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(time.Minute*5).String())
	})
	t.Run("TestBegin2", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:55:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(time.Minute*5).String())
	})
	t.Run("TestBegin3", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:54:59")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(time.Minute*5).String())
	})

	t.Run("TestEnd1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:01:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd2", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd3", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 4:56:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), end.String())
	})

	t.Run("TestLoop1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:46:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		var last time.Time

		for i := 0; i < 1000; i++ {
			next := sch.Next(current)
			if next.IsZero() {
				break
			}
			last = next
			current = next
		}
		require.Equal(t, last.String(), end.String())
	})
}

func TestSchedule_UnixCron2(t *testing.T) {
	var err error
	var begin time.Time
	var end time.Time

	begin, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:00:01")
	require.Nil(t, err)
	end, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:02")
	require.Nil(t, err)

	begin = begin.In(time.Local)
	end = end.In(time.Local)

	sch := UnixCron{
		Begin:        begin,
		End:          end,
		Express:      "*/5 * * * *",
		once:         sync.Once{},
		exprSchedule: nil,
	}
	t.Run("TestBegin1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:00:01")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(time.Minute*5).Add(-time.Second*1).String())
	})

	t.Run("TestEnd1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd2", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:01")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd3", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 4:56:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), end.Add(-time.Second*2).String())
	})

	t.Run("TestLoop1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:46:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		var last time.Time

		for i := 0; i < 1000; i++ {
			next := sch.Next(current)
			if next.IsZero() {
				break
			}
			last = next
			current = next
		}
		require.Equal(t, last.String(), end.Add(-time.Second*2).String())
	})
}

// TestSchedule_UnixCron3 for test next time before `Begin` time,
func TestSchedule_UnixCron3(t *testing.T) {
	// Time string is UTC.
	beginTimeSeeds := map[string]time.Duration{
		"2022-01-18 3:05:00": time.Minute * 5,
		"2022-01-18 3:05:01": time.Minute*5 - time.Second,
		"2022-01-18 3:06:00": time.Minute * 4,
	}

	// Time string is UTC.
	currentTimeSeeds := []string{
		"1997-01-18 3:00:00",
		"2022-01-18 2:00:00",
		"2022-01-18 2:56:00",
		"2022-01-18 3:00:00",
		"2022-01-18 3:04:00",
		"2022-01-18 3:05:00",
		"2022-01-18 3:05:01",
		"2022-01-18 3:09:55",
	}

	for beginTimeStr, delay := range beginTimeSeeds {
		begin, err := time.Parse("2006-01-02 15:04:05", beginTimeStr)
		require.Nil(t, err)
		begin = begin.In(time.Local)
		sch := UnixCron{
			Begin:        begin,
			End:          time.Time{},
			Express:      "*/5 * * * *",
			once:         sync.Once{},
			exprSchedule: nil,
		}
		for _, currentTimeStr := range currentTimeSeeds {
			current, err := time.Parse("2006-01-02 15:04:05", currentTimeStr)
			require.Nil(t, err)
			current = current.In(time.Local)
			next := sch.Next(current)
			msg := fmt.Sprintf("beging: %s, current: %s", beginTimeStr, currentTimeStr)
			require.Equal(t, next.String(), begin.Add(delay).String(), msg)
		}
	}
}

func TestSchedule_BeginGreaterThanEnd(t *testing.T) {
	var err error
	var begin time.Time
	var end time.Time

	begin, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 6:00:00")
	require.Nil(t, err)
	end, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
	require.Nil(t, err)

	begin = begin.In(time.Local)
	end = end.In(time.Local)

	t.Run("UnixCron", func(t *testing.T) {
		sch := UnixCron{
			Begin:        begin,
			End:          end,
			Express:      "*/5 * * * *",
			once:         sync.Once{},
			exprSchedule: nil,
		}

		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:05:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), time.Time{}.String())
	})

	t.Run("Interval", func(t *testing.T) {
		sch := Interval{
			Begin:    begin,
			End:      end,
			Interval: time.Second,
		}

		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:05:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), time.Time{}.String())
	})
}

func TestSchedule_Interval1(t *testing.T) {
	var err error
	var begin time.Time
	var end time.Time

	begin, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:00:00")
	require.Nil(t, err)
	end, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
	require.Nil(t, err)

	begin = begin.In(time.Local)
	end = end.In(time.Local)

	sch := Interval{
		Begin:    begin,
		End:      end,
		Interval: time.Second,
	}

	t.Run("TestNext1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:05:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), current.Add(time.Second*1).String())
	})

	t.Run("TestBegin1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:00:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(sch.Interval).String())
	})
	t.Run("TestBegin2", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:55:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(sch.Interval).String())
	})
	t.Run("TestBegin3", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 2:54:59")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(sch.Interval).String())
	})

	t.Run("TestEnd1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:01:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd2", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 5:00:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.True(t, next.IsZero())
	})

	t.Run("TestEnd3", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 4:56:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), current.Add(time.Second).String())
	})

	t.Run("TestLoop1", func(t *testing.T) {
		var current time.Time
		current, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:46:00")
		require.Nil(t, err)
		current = current.In(time.Local)

		var last time.Time

		for i := 0; i < 1000000; i++ {
			next := sch.Next(current)
			if next.IsZero() {
				break
			}
			last = next
			current = next
		}
		require.Equal(t, last.String(), end.String())
	})
}

func TestSchedule_Interval2(t *testing.T) {
	var err error
	var begin time.Time

	begin, err = time.Parse("2006-01-02 15:04:05", "2022-01-18 3:00:00")
	require.Nil(t, err)
	begin = begin.In(time.Local)

	sch := Interval{
		Begin:    begin,
		End:      time.Time{},
		Interval: time.Minute,
	}

	seedTimes := []string{
		"2022-01-18 2:05:00",
		"2022-01-18 2:55:01",
		"2022-01-18 3:00:00",
	}

	var current time.Time
	for _, timeStr := range seedTimes {
		current, err = time.Parse("2006-01-02 15:04:05", timeStr)
		require.Nil(t, err)
		current = current.In(time.Local)

		next := sch.Next(current)
		require.Equal(t, next.String(), begin.Add(sch.Interval).String())
	}
}
