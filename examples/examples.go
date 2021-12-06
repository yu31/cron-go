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

	_ = c.Once(
		&gcron.Task{
			Ctx:   context.Background(),
			Key:   "once1",
			Value: "1024",
			Callback: func(ctx context.Context, key string, value interface{}) error {
				fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Once{Time: time.Now().Add(time.Second)},
	)

	_ = c.Express(
		&gcron.Task{
			Ctx:   context.Background(),
			Key:   "express1",
			Value: nil,
			Callback: func(ctx context.Context, key string, value interface{}) error {
				fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Express{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	_ = c.Express(
		&gcron.Task{
			Ctx:   context.Background(),
			Key:   "express2",
			Value: nil,
			Callback: func(ctx context.Context, key string, value interface{}) error {
				fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Express{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	_ = c.Interval(
		&gcron.Task{
			Ctx:   context.Background(),
			Key:   "interval1",
			Value: nil,
			Callback: func(ctx context.Context, key string, value interface{}) error {
				fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Interval{
			Begin:    time.Unix(662688000, 0),
			End:      time.Unix(2556144000, 0),
			Interval: time.Second,
		},
	)

	_ = c.Interval(
		&gcron.Task{
			Ctx:   context.Background(),
			Key:   "interval1",
			Value: nil,
			Callback: func(ctx context.Context, key string, value interface{}) error {
				fmt.Println("run jod:", "key:", key, "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Interval{
			Begin:    time.Unix(662688000, 0),
			End:      time.Unix(2556144000, 0),
			Interval: time.Second * 3,
		},
	)

	time.Sleep(time.Second * 120)
}
