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
		Time:  time.Now().Add(time.Second),
		Ctx:   context.Background(),
		Value: "1024",
		Callback: func(ctx context.Context, key string, value interface{}) error {
			fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
			return nil
		},
	})

	_ = c.Express("express1", &gcron.Express{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Express:     "* * * * *",
		Ctx:         context.Background(),
		Value:       nil,
		Callback: func(ctx context.Context, key string, value interface{}) error {
			fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
			return nil
		},
	})

	_ = c.Express("express2", &gcron.Express{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Express:     "* * * * *",
		Ctx:         context.Background(),
		Value:       nil,
		Callback: func(ctx context.Context, key string, value interface{}) error {
			fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
			return nil
		},
	})

	_ = c.Interval("interval1", &gcron.Interval{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Ctx:         context.Background(),
		Interval:    time.Second,
		Value:       nil,
		Callback: func(ctx context.Context, key string, value interface{}) error {
			fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
			return nil
		},
	})

	_ = c.Interval("interval2", &gcron.Interval{
		Begin:       time.Unix(662688000, 0),
		End:         time.Unix(2556144000, 0),
		Concurrency: false,
		Ctx:         context.Background(),
		Interval:    time.Second * 3,
		Value:       nil,
		Callback: func(ctx context.Context, key string, value interface{}) error {
			fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
			return nil
		},
	})

	time.Sleep(time.Second * 120)
}
