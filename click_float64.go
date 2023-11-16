package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-vgo/robotgo"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

/**
	* 实现自动点击功能
	* 命令示例： click.exe 2 60   间隔点击时间2s, 超时关闭时间60s
	* f9 开启自动点击, f10 关闭自动点击
**/
func main() {
	err := disableQuickEditMode()
	if err != nil {
		fmt.Printf("禁用快速编辑模式失败： %v\n", err)
		return
	}
	fmt.Println("当前窗口已禁用快速编辑模式")

	args := os.Args
	config, err := parseInputArgs(args)
	if err != nil {
		fmt.Println("参数解析错误:", err)
		return
	}
	interval := config.Interval // 默认间隔时间，单位为秒
	timeout := config.Timeout   // 默认超时关闭时间，单位为秒

	idleTimeout := time.NewTimer(time.Duration(timeout) * time.Second)

	open := "f9"
	close := "f10"
	fmt.Println("已设置点击间隔时间为", interval, "秒")
	fmt.Println("已设置自动关闭时间为", timeout, "秒")
	fmt.Printf("按下 %s 启动自动点击，按下 %s 停止自动点击\n", strings.ToUpper(open), strings.ToUpper(close))

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
						resetIdleTimeout(idleTimeout, time.Duration(timeout)*time.Second)
						// 休眠
						time.Sleep(time.Duration(interval * float64(time.Second)))
					}
				}()
			}
		case <-stopAutoClick:
			if autoClicking {
				autoClicking = false
				fmt.Println("停止自动点击")
				// 重置自动关闭计时器
				resetIdleTimeout(idleTimeout, time.Duration(timeout)*time.Second)
			}
		case sig := <-getOSInterruptChannel():
			fmt.Println("接收到信号:", sig)
			autoClicking = false
			return
		}
	}
}

// 参数配置
type Config struct {
	Interval float64
	Timeout  int
}

func parseInputArgs(args []string) (Config, error) {
	var config Config
	config.Interval = 2.0
	config.Timeout = 1800

	if len(args) > 1 {
		argInterval := args[1]
		parsedInterval, err := time.ParseDuration(argInterval + "s")
		if err == nil {
			config.Interval = parsedInterval.Seconds()
		} else {
			fmt.Println("无效的间隔时间参数，默认使用", config.Interval, "秒")
		}
		if len(args) > 2 {
			timeInterval := args[2]
			parsedTimeout, err := time.ParseDuration(timeInterval + "s")
			if err != nil {
				fmt.Println("无效的自动关闭时间参数，默认使用", config.Timeout, "秒")
			} else {
				config.Timeout = int(parsedTimeout.Seconds())
			}
		}
	} else {
		var inputInTerval float64
		var inputTimeout int
		fmt.Println("请输入点击间隔时间(秒)，支持小数，不输入默认为", config.Interval, "秒：")
		_, err := fmt.Scanln(&inputInTerval)
		if err == nil {
			config.Interval = inputInTerval
		}
		fmt.Println("请输入自动关闭时间(秒)，不输入默认为", config.Timeout, "秒：")
		_, err = fmt.Scanln(&inputTimeout)
		if err == nil {
			config.Timeout = inputTimeout
		}
	}
	if config.Interval < 0.1 {
		config.Interval = 0.1
	}
	return config, nil
}

func getOSInterruptChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}

func resetIdleTimeout(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		<-timer.C
	}
	timer.Reset(duration)
}

func disableQuickEditMode() error {
	var mode uint32
	stdoutHandle := syscall.Stdout

	// 获取当前控制台模式
	ret, _, err := procGetConsoleMode.Call(uintptr(stdoutHandle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return fmt.Errorf("获取控制台模式失败: %v", err)
	}

	// 取消快速编辑模式
	const ENABLE_QUICK_EDIT_MODE uint32 = 0x40
	newMode := mode &^ ENABLE_QUICK_EDIT_MODE

	// 设置新的控制台模式
	ret, _, err = procSetConsoleMode.Call(uintptr(stdoutHandle), uintptr(newMode))
	if ret == 0 {
		return fmt.Errorf("设置控制台模式失败: %v", err)
	}

	return nil
}
