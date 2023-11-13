package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)
/**
	* V2版本:实现自动点击功能
	* 命令示例： click.exe 2 60   间隔点击时间2s, 超时关闭时间60s
	* f9 开启自动点击, f10 关闭自动点击
**/
func main() {
	args := os.Args
	interval := 2 // 默认间隔时间，单位为秒
	timeout := 180// 默认超时关闭时间，单位为秒

	if len(args) > 1 {
		argInterval := args[1]
		parsedInterval, err := time.ParseDuration(argInterval + "s")
		if err == nil {
			interval = int(parsedInterval.Seconds())
		} else {
			fmt.Println("无效的间隔时间参数，默认使用", interval, "秒")
		}
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
	idleTimeout := time.NewTimer(time.Duration(timeout) * time.Second)

	open := "f9"
	close := "f10"
	fmt.Println("已设置点击间隔时间为", interval, "秒")
	fmt.Println("已设置自动关闭时间为", timeout, "秒")
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
		case <-idleTimeout.C:
			fmt.Println("程序空闲超过", timeout, "秒，自动退出")
			return

		case <-startAutoClick:
			if !autoClicking {
				autoClicking = true
				fmt.Println("开始自动点击")
				go func() {
					for autoClicking {
						robotgo.MouseClick("left", false)
						// 重置自动关闭计时器
						resetIdleTimeout(idleTimeout, time.Duration(timeout) * time.Second)
						// 休眠
						time.Sleep(time.Duration(interval) * time.Second)
					}
				}()
			}
		case <-stopAutoClick:
			if autoClicking {
				autoClicking = false
				fmt.Println("停止自动点击")
				// 重置自动关闭计时器
				resetIdleTimeout(idleTimeout, time.Duration(timeout) * time.Second)
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


func resetIdleTimeout(timer * time.Timer, duration time.Duration) {
	if !timer.Stop() {
		<-timer.C
	}
	timer.Reset(duration)
}
