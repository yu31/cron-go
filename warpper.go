package gcron

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type ScheduleWrapper func(s Schedule) Schedule

type wrapperRecover struct {
	Schedule
}

func (w *wrapperRecover) Run() {
	if r := recover(); r != nil {
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}
		// FIXME:
		println(fmt.Sprintf("gcron: panic with error: %v, stack: \n%s", err, string(buf)))
	}
	w.Schedule.Run()
}

type wrapperWaitGroup struct {
	Schedule
	wg *sync.WaitGroup
}

func (w *wrapperWaitGroup) Run() {
	w.wg.Add(1)
	w.Schedule.Run()
	w.wg.Done()
}

type wrapperRetry struct {
	Schedule
	ctx      context.Context
	limit    int64         // max retry limit.
	interval time.Duration // retry interval
}

func (w *wrapperRetry) Run() {
	w.Schedule.Run()
	if w.Schedule.Err() == nil {
		return
	}
	if w.limit <= 0 {
		return
	}

	ticker := time.NewTicker(w.interval)
LOOP:
	for i := int64(0); i < w.limit; i++ {
		select {
		case <-ticker.C:
			w.Schedule.Run()
		case <-w.ctx.Done():
			break LOOP
		case <-w.Schedule.Context().Done():
			break LOOP
		}
		if w.Schedule.Err() == nil {
			break LOOP
		}
	}
	ticker.Stop()
}
