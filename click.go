package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	args := os.Args
	interval := 5 // 默认间隔时间，单位为秒
	if len(args) > 1 {
		argInterval := args[1]
		fmt.Println("从命令行设置点击间隔时间为", argInterval, "秒")
		parsedInterval, err := time.ParseDuration(argInterval + "s")
		if err == nil {
			interval = int(parsedInterval.Seconds())
		} else {
			fmt.Println("无效的间隔时间参数，默认使用", interval, "秒")
		}
	}
	open := "f9"
	close := "f10"
	fmt.Printf("按下 %s 启动自动点击，按下 %s 停止自动点击\n", open, close)

	startAutoClick := make(chan bool)
	stopAutoClick := make(chan bool)

	go func() {
		for {
			if robotgo.AddEvent(open) {
				startAutoClick <- true
			}
			if robotgo.AddEvent(close) {
				stopAutoClick <- true
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	autoClicking := false
	for {
		select {
		case <-startAutoClick:
			if !autoClicking {
				autoClicking = true
				fmt.Println("开始自动点击")
				go func() {
					for autoClicking {
						robotgo.MouseClick("left", false)
						time.Sleep(time.Duration(interval) * time.Second)
					}
				}()
			}
		case <-stopAutoClick:
			if autoClicking {
				autoClicking = false
				fmt.Println("停止自动点击")
			}
		case sig := <-getOSInterruptChannel():
			fmt.Println("接收到信号:", sig)
			autoClicking = false
			return
		}
	}
}

func getOSInterruptChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
