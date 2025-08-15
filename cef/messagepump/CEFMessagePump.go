package messagepump

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include <Cocoa/Cocoa.h>
#include <pthread.h>
#include <stdlib.h>

typedef void (*GoFunc)(void);

// 主线程RunLoop
static CFRunLoopRef mainRunLoop;

// 初始化主线程RunLoop
static void init_main_runloop() {
    mainRunLoop = CFRunLoopGetMain();
}

// 任务结构体
typedef struct {
    GoFunc func;
} CEFTask;

// 在主线程执行任务
static void perform_task(void* info) {
    CEFTask* task = (CEFTask*)info;
    if (task->func) {
        task->func();
    }
    free(task);
}

// 将任务投递到主线程
static void post_to_main_thread(GoFunc func) {
    CEFTask* task = (CEFTask*)malloc(sizeof(CEFTask));
    task->func = func;

    CFRunLoopPerformBlock(mainRunLoop, kCFRunLoopCommonModes, ^{
        perform_task(task);
    });
}

static void CFRunLoopWakeUpMainRunLoop() {
    CFRunLoopWakeUp(mainRunLoop);
}

// 获取当前线程ID
static uint64_t get_current_thread_id() {
    return pthread_mach_thread_np(pthread_self());
}

// 关键修复：声明Go导出的函数
extern void doMessageLoopWorkCallback();
*/
import "C"
import (
	"github.com/energye/cef/cef"
	"github.com/energye/lcl/api"
	"sync"
	"time"
)

// CEFMessagePump 管理CEF消息循环的控制器
type CEFMessagePump struct {
	mu           sync.Mutex
	timer        *time.Timer
	scheduleCh   chan int64
	quitCh       chan struct{}
	mainThreadID uint64
	running      bool
}

var (
	globalPump   *CEFMessagePump
	GlobalCEFApp cef.ICefApplication
)

// InitMessagePump 初始化消息泵（必须在主线程调用）
func InitMessagePump() {
	// 初始化主线程RunLoop
	C.init_main_runloop()

	globalPump = &CEFMessagePump{
		scheduleCh:   make(chan int64, 100),
		quitCh:       make(chan struct{}),
		mainThreadID: uint64(C.get_current_thread_id()),
	}

	// 启动调度处理goroutine
	go globalPump.processSchedules()
}

// 处理调度事件
func (p *CEFMessagePump) processSchedules() {
	p.running = true
	defer func() { p.running = false }()

	for {
		select {
		case delayMs := <-p.scheduleCh:
			p.handleSchedule(delayMs)
		case <-p.quitCh:
			return
		}
	}
}

// 处理调度逻辑
func (p *CEFMessagePump) handleSchedule(delayMs int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 取消当前定时器
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}

	if delayMs <= 0 {
		// 立即执行
		p.postToMainThread()
	} else {
		// 延迟执行
		p.timer = time.AfterFunc(
			time.Duration(delayMs)*time.Millisecond,
			p.postToMainThread,
		)
	}
}

// 将任务投递到主线程
func (p *CEFMessagePump) postToMainThread() {
	// 直接使用导出的C函数
	C.post_to_main_thread(C.GoFunc(C.doMessageLoopWorkCallback))
	C.CFRunLoopWakeUpMainRunLoop()
}

//export doMessageLoopWorkCallback
func doMessageLoopWorkCallback() {
	println("  ✅ CEF消息循环工作执行成功! 当前线程ID:", C.get_current_thread_id(), api.CurrentThreadId(),
		"主线程ID:", globalPump.mainThreadID, api.MainThreadId())
	GlobalCEFApp.DoMessageLoopWork()
	println("  ✅ CEF消息循环工作执行结束")
}

// OnScheduleMessagePumpWork 处理CEF的调度回调
func OnScheduleMessagePumpWork(delayMs int64) {
	if globalPump != nil && globalPump.running {
		globalPump.scheduleCh <- delayMs
	}
}

// Shutdown 关闭消息泵和CEF
func Shutdown() {
	if globalPump != nil {
		close(globalPump.quitCh)
		globalPump.mu.Lock()
		if globalPump.timer != nil {
			globalPump.timer.Stop()
		}
		globalPump.mu.Unlock()
	}
}

// IsMainThread 验证当前是否为主线程
func IsMainThread() bool {
	if globalPump == nil {
		return false
	}
	return uint64(C.get_current_thread_id()) == globalPump.mainThreadID
}
