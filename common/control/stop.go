package control

import (
	"bygame/common/log"
	"os"
	"os/signal"
	"sort"
	"syscall"
)

type stopFuncData struct {
	f        func() // 要执行的操作
	priority int    // 优先级
}

var stopFunc = make([]stopFuncData, 0)
var exitSignal = make(chan os.Signal)

// 阻塞进程
func run() {
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)
	<-exitSignal
	log.Inf("收到关服指令")
	stop()
	sort.Slice(stopFunc, func(i, j int) bool {
		return stopFunc[i].priority > stopFunc[j].priority
	})
	for _, sfd := range stopFunc {
		sfd.f()
	}
	remove()
	log.Flush()
}

// 手动关闭服务器
func Stop() {
	exitSignal <- os.Interrupt
}

// 添加一个关闭前操作
// priority 优先级数字越大越先执行
func AddStopFunc(f func(), priority int) {
	stopFunc = append(stopFunc, stopFuncData{f, priority})
}
