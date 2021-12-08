package main

import (
	"fmt"
	"time"

	"github.com/yu31/gcron"
)

func main() {
	cron := gcron.New()
	cron.Start()
	defer cron.Stop()

	cron.Submit(
		"once1",
		&gcron.Task{
			Value: "1024",
			Callback: func(value interface{}) error {
				fmt.Println("run jod:", "key: once1", "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Once{Time: time.Now().Add(time.Second)},
	)

	cron.Submit(
		"unix_cron1",
		&gcron.Task{
			Value: nil,
			Callback: func(value interface{}) error {
				fmt.Println("run jod:", "key: unix_cron1", "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.UnixCron{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	cron.Submit(
		"unix_cron2",
		&gcron.Task{
			Value: nil,
			Callback: func(value interface{}) error {
				fmt.Println("run jod:", "key: unix_cron2", "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.UnixCron{
			Begin:   time.Unix(662688000, 0),
			End:     time.Unix(2556144000, 0),
			Express: "* * * * *",
		},
	)

	cron.Submit(
		"interval1",
		&gcron.Task{
			Value: nil,
			Callback: func(value interface{}) error {
				fmt.Println("run jod:", "key: interval1", "value:", value, time.Now().String())
				return nil
			},
		},
		&gcron.Interval{
			Begin:    time.Unix(662688000, 0),
			End:      time.Unix(2556144000, 0),
			Interval: time.Second,
		},
	)

	cron.Submit(
		"interval2",
		&gcron.Task{
			Value: nil,
			Callback: func(value interface{}) error {
				fmt.Println("run jod:", "key: interval2", "value:", value, time.Now().String())
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
