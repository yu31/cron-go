package gcron

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerLog "github.com/opentracing/opentracing-go/log"
)

type wrapperRecover struct {
	Job
}

func (w *wrapperRecover) Run() (err error) {
	if r := recover(); r != nil {
		var ok bool
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		err, ok = r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}
		// FIXME:
		println(fmt.Sprintf("gcron: panic with error: %v, stack: \n%s", err, string(buf)))
	}
	err = w.Job.Run()
	return
}

type wrapperWaitGroup struct {
	Job
	wg *sync.WaitGroup
}

func (w *wrapperWaitGroup) Run() (err error) {
	w.wg.Add(1)
	err = w.Job.Run()
	w.wg.Done()
	return
}

type wrapperRetry struct {
	Job
	ctx      context.Context
	limit    int64         // max retry limit.
	interval time.Duration // retry interval
}

func (w *wrapperRetry) Run() (err error) {
	err = w.Job.Run()
	if err == nil {
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
			if err = w.Job.Run(); err == nil {
				break LOOP
			}
		case <-w.ctx.Done():
			break LOOP
		case <-w.Job.Context().Done():
			break LOOP
		}
	}
	ticker.Stop()

	return
}

type wrapperSkipIfRunning struct {
	Job
	running int32 // running indicates whether the task func is running. 1 => true, 0 => false.
}

func (w *wrapperSkipIfRunning) Run() (err error) {
	// To avoids run repeat.
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return
	}
	err = w.Job.Run()
	atomic.StoreInt32(&w.running, 0)
	return
}

type wrapperDelayIfRunning struct {
	Job
	mu sync.Mutex
}

func (w *wrapperDelayIfRunning) Run() (err error) {
	w.mu.Lock()
	err = w.Job.Run()
	w.mu.Unlock()
	return
}

type wrapperOpentracing struct {
	Job
	tracer opentracing.Tracer
}

func (w *wrapperOpentracing) Run() (err error) {
	ctx := w.Job.Context()

	var parentCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx = parent.Context()
	}

	span := w.tracer.StartSpan(
		"GCronRun",
		opentracing.ChildOf(parentCtx),
		opentracing.Tag{Key: string(ext.Component), Value: "GCron"},
	)
	span.LogFields(tracerLog.String("key", w.Job.TaskKey()))

	ctx = opentracing.ContextWithSpan(ctx, span)

	w.Job.WithContext(ctx)

	if err = w.Job.Run(); err != nil {
		span.LogFields(tracerLog.Error(err))
	}

	span.Finish()
	return
}
