package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yu31/gcron"
)

func main() {
	c := gcron.New()
	c.Start()
	defer c.Stop()

	_ = c.Once("once1", &gcron.Once{
		Time:    time.Now().Add(time.Second),
		Context: context.Background(),
		Value:   "1024",
		Callback: func(ctx context.Context, value interface{}) {
			fmt.Println("run jod once1,", value, time.Now().String())
		},
	})

	_ = c.Express("express1", &gcron.Express{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Express:     "* * * * *",
		Context:     context.Background(),
		Value:       nil,
		Callback: func(ctx context.Context, value interface{}) {
			fmt.Println("run jod express1,", time.Now().String())
		},
	})

	_ = c.Express("express2", &gcron.Express{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Express:     "* * * * *",
		Context:     context.Background(),
		Value:       nil,
		Callback: func(ctx context.Context, value interface{}) {
			fmt.Println("run jod express2,", time.Now().String())
		},
	})

	_ = c.Interval("interval1", &gcron.Interval{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Context:     context.Background(),
		Interval:    time.Second,
		Value:       nil,
		Callback: func(ctx context.Context, value interface{}) {
			fmt.Println("run jod interval1,", time.Now().String())
		},
	})

	_ = c.Interval("interval2", &gcron.Interval{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Context:     context.Background(),
		Interval:    time.Second * 3,
		Value:       nil,
		Callback: func(ctx context.Context, value interface{}) {
			fmt.Println("run jod interval2,", time.Now().String())
		},
	})

	time.Sleep(time.Second * 120)
}
