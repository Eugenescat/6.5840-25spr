package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// 互斥锁保护共享数据
	mu sync.Mutex

	// Map任务相关
	mapTasks []Task
	mapDone  bool
	mapFiles []string

	// Reduce任务相关
	reduceTasks []Task
	reduceDone  bool

	// 任务配置
	nReduce int
	nMap    int
}

// Your code here -- RPC handlers for the worker to call.

// RequestTask RPC处理函数，处理worker请求任务的请求
func (c *Coordinator) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查Map阶段是否完成
	if !c.mapDone {
		// 分配Map任务
		for i := range c.mapTasks {
			if c.mapTasks[i].Status == Idle {
				c.mapTasks[i].Status = InProgress
				c.mapTasks[i].StartTime = time.Now()
				reply.Task = c.mapTasks[i]
				return nil
			}
		}
		// 检查是否有超时的Map任务
		for i := range c.mapTasks {
			if c.mapTasks[i].Status == InProgress && time.Since(c.mapTasks[i].StartTime) > 10*time.Second {
				c.mapTasks[i].StartTime = time.Now()
				reply.Task = c.mapTasks[i]
				return nil
			}
		}
		// 如果没有可用的Map任务，返回NoTask
		reply.Task = Task{Type: NoTask}
		return nil
	}

	// Map阶段已完成，检查Reduce阶段
	if !c.reduceDone {
		// 分配Reduce任务
		for i := range c.reduceTasks {
			if c.reduceTasks[i].Status == Idle {
				c.reduceTasks[i].Status = InProgress
				c.reduceTasks[i].StartTime = time.Now()
				reply.Task = c.reduceTasks[i]
				return nil
			}
		}
		// 检查是否有超时的Reduce任务
		for i := range c.reduceTasks {
			if c.reduceTasks[i].Status == InProgress && time.Since(c.reduceTasks[i].StartTime) > 10*time.Second {
				c.reduceTasks[i].StartTime = time.Now()
				reply.Task = c.reduceTasks[i]
				return nil
			}
		}
		// 如果没有可用的Reduce任务，返回NoTask
		reply.Task = Task{Type: NoTask}
		return nil
	}

	// 所有任务都完成了，返回退出信号
	reply.Task = Task{Type: ExitTask}
	return nil
}

// ReportTask RPC处理函数，处理worker报告任务完成的请求
func (c *Coordinator) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if args.TaskType == MapTask {
		if args.TaskId >= 0 && args.TaskId < len(c.mapTasks) {
			c.mapTasks[args.TaskId].Status = Completed
			// 检查是否所有Map任务都完成了
			c.checkMapDone()
		}
	} else if args.TaskType == ReduceTask {
		if args.TaskId >= 0 && args.TaskId < len(c.reduceTasks) {
			c.reduceTasks[args.TaskId].Status = Completed
			// 检查是否所有Reduce任务都完成了
			c.checkReduceDone()
		}
	}
	return nil
}

// checkMapDone 检查所有Map任务是否完成
func (c *Coordinator) checkMapDone() {
	for _, task := range c.mapTasks {
		if task.Status != Completed {
			return
		}
	}
	c.mapDone = true
}

// checkReduceDone 检查所有Reduce任务是否完成
func (c *Coordinator) checkReduceDone() {
	for _, task := range c.reduceTasks {
		if task.Status != Completed {
			return
		}
	}
	c.reduceDone = true
}

// Done 检查整个MapReduce作业是否完成
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mapDone && c.reduceDone
}

// MakeCoordinator 创建一个新的Coordinator
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		mapFiles:    files,
		nReduce:     nReduce,
		nMap:        len(files),
		mapTasks:    make([]Task, len(files)),
		reduceTasks: make([]Task, nReduce),
		mapDone:     false,
		reduceDone:  false,
	}

	// 初始化Map任务
	for i := range c.mapTasks {
		c.mapTasks[i] = Task{
			Type:     MapTask,
			Id:       i,
			Status:   Idle,
			FileName: files[i],
			NReduce:  nReduce,
			NMap:     len(files),
		}
	}

	// 初始化Reduce任务
	for i := range c.reduceTasks {
		c.reduceTasks[i] = Task{
			Type:    ReduceTask,
			Id:      i,
			Status:  Idle,
			NReduce: nReduce,
			NMap:    len(files),
		}
	}

	c.server()
	return &c
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}
