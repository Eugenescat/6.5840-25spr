package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"time"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// TaskType 定义任务类型
type TaskType int

const (
	MapTask TaskType = iota
	ReduceTask
	NoTask   // 没有可用任务
	ExitTask // 退出信号
)

// TaskStatus 定义任务状态
type TaskStatus int

const (
	Idle       TaskStatus = iota // 任务未分配
	InProgress                   // 任务正在执行
	Completed                    // 任务已完成
)

// Task 定义任务结构
type Task struct {
	Type      TaskType   // 任务类型
	Id        int        // 任务ID
	Status    TaskStatus // 任务状态
	FileName  string     // 对于Map任务，这是输入文件名
	NReduce   int        // reduce任务数量
	NMap      int        // map任务数量
	StartTime time.Time  // 任务开始时间
}

// RequestTaskArgs Worker请求任务时的参数
type RequestTaskArgs struct {
	// 可以为空，因为Worker只需要表明它准备好了
}

// RequestTaskReply Coordinator回复任务请求的结构
type RequestTaskReply struct {
	Task Task
}

// ReportTaskArgs Worker报告任务完成时的参数
type ReportTaskArgs struct {
	TaskId   int
	TaskType TaskType
}

// ReportTaskReply Coordinator回复任务报告的结构
type ReportTaskReply struct {
	// 可以为空，因为Worker只需要知道报告是否成功
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
