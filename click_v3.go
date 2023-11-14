package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	interval, timeout := parseArgs(os.Args)

	idleTimeout := time.NewTimer(time.Duration(timeout) * time.Second)
	defer idleTimeout.Stop()

	open := "f9"
	close := "f10"

	fmt.Println("已设置点击间隔时间为", interval, "秒")
	fmt.Println("已设置自动关闭时间为", timeout, "秒")
	fmt.Printf("按下 %s 启动自动点击，按下 %s 停止自动点击\n", strings.ToUpper(open), strings.ToUpper(close))

	startAutoClick := make(chan bool)
	stopAutoClick := make(chan bool)

	go listenForEvents(open, close, startAutoClick, stopAutoClick)

	handleAutoClick(idleTimeout, interval, timeout, startAutoClick, stopAutoClick)
}

/*
parseArgs 解析命令行参数并返回间隔时间和超时时间。

args: 包含命令行参数的字符串切片。
返回两个整数：间隔时间和超时时间。
*/
func parseArgs(args []string) (int, int) {
	interval := 2
	timeout := 1800

	if len(args) > 1 {
		argInterval := args[1]
		parsedInterval, err := time.ParseDuration(argInterval + "s")
		if err == nil {
			interval = int(parsedInterval.Seconds())
		} else {
			fmt.Println("无效的间隔时间参数，默认使用", interval, "秒")
		}
		if len(args) > 2 {
			timeInterval := args[2]
			parsedTimeout, err := time.ParseDuration(timeInterval + "s")
			if err != nil {
				fmt.Println("无效的自动关闭时间参数，默认使用", timeout, "秒")
			} else {
				timeout = int(parsedTimeout.Seconds())
			}
		}
	} else {
		var inputInTerval, inputTimeout int
		fmt.Println("请输入点击间隔时间(秒)，不输入默认为", interval, "秒：")
		_, err := fmt.Scanln(&inputInTerval)
		if err == nil {
			interval = inputInTerval
		}
		fmt.Println("请输入自动关闭时间(秒)，不输入默认为", timeout, "秒：")
		_, err = fmt.Scanln(&inputTimeout)
		if err == nil {
			timeout = inputTimeout
		}
	}

	return interval, timeout
}

func listenForEvents(open string, close string, startAutoClick chan<- bool, stopAutoClick chan<- bool) {
	if robotgo.AddEvent(open) {
		startAutoClick <- true
	}
	if robotgo.AddEvent(close) {
		stopAutoClick <- true
	}
	time.Sleep(100 * time.Millisecond)
}

func handleAutoClick(idleTimeout *time.Timer, interval int, timeout int, startAutoClick <-chan bool, stopAutoClick <-chan bool) {
	autoClick := false

	for {
		select {
		case <-idleTimeout.C:
			fmt.Println("程序空闲超过", timeout, "秒，自动退出")
			return

		case <-startAutoClick:
			if !autoClick {
				autoClick = true
				fmt.Println("开始自动点击")
				go autoClicking(interval, idleTimeout, timeout, &autoClick)
			}

		case <-stopAutoClick:
			if autoClick {
				autoClick = false
				fmt.Println("停止自动点击")
				resetIdleTimeout(idleTimeout, time.Duration(timeout)*time.Second)
			}

		case sig := <-getOSInterruptChannel():
			fmt.Println("接收到信号：", sig)
			autoClick = false
			return
		}
	}
}

func autoClicking(interval int, idleTimeout *time.Timer, timeout int, autoClick *bool) {
	for *autoClick {
		robotgo.Click()
		resetIdleTimeout(idleTimeout, time.Duration(timeout)*time.Second)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func resetIdleTimeout(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

func getOSInterruptChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
