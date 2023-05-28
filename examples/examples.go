package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yu31/cron-go"
)

func main() {
	crontab := cron.New()
	crontab.Start()
	defer crontab.Stop()

	crontab.Submit(
		context.Background(),
		"once1",
		&cron.Task{
			Value: "1024",
			Callback: func(ctx context.Context, value interface{}) error {
				fmt.Println("run jod:", "key: once1", "value:", value, time.Now().String())
				return nil
			},
		},
		&cron.Appoint{Time: time.Now().Add(time.Second)},
	)

	crontab.Submit(
		context.Background(),
		"unix_cron1",
		&cron.Task{
			Value: nil,
			Callback: func(ctx context.Context, value interface{}) error {
				fmt.Println("run jod:", "key: unix_cron1", "value:", value, time.Now().String())
				return nil
			},
		},
		&cron.UnixCron{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	crontab.Submit(
		context.Background(),
		"unix_cron2",
		&cron.Task{
			Value: nil,
			Callback: func(ctx context.Context, value interface{}) error {
				fmt.Println("run jod:", "key: unix_cron2", "value:", value, time.Now().String())
				return nil
			},
		},
		&cron.UnixCron{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	crontab.Submit(
		context.Background(),
		"interval1",
		&cron.Task{
			Value: nil,
			Callback: func(ctx context.Context, value interface{}) error {
				fmt.Println("run jod:", "key: interval1", "value:", value, time.Now().String())
				return nil
			},
		},
		&cron.Interval{
			Begin:    time.Unix(662688000, 0),
			End:      time.Unix(2556144000, 0),
			Interval: time.Second,
		},
	)

	crontab.Submit(
		context.Background(),
		"interval2",
		&cron.Task{
			Value: nil,
			Callback: func(ctx context.Context, value interface{}) error {
				fmt.Println("run jod:", "key: interval2", "value:", value, time.Now().String())
				return nil
			},
		},
		&cron.Interval{
			Begin:    time.Unix(662688000, 0),
			End:      time.Unix(2556144000, 0),
			Interval: time.Second * 3,
		},
	)

	time.Sleep(time.Second * 120)
}
