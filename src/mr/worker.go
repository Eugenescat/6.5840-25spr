package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// Worker 实现MapReduce的worker逻辑
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	for {
		// 请求任务
		args := RequestTaskArgs{}
		reply := RequestTaskReply{}

		ok := call("Coordinator.RequestTask", &args, &reply)
		if !ok {
			// 如果无法连接到coordinator，认为coordinator已经退出，worker也退出
			return
		}

		task := reply.Task
		switch task.Type {
		case MapTask:
			doMapTask(task, mapf)
		case ReduceTask:
			doReduceTask(task, reducef)
		case NoTask:
			// 没有可用任务，等待一段时间后重试
			time.Sleep(time.Second)
			continue
		case ExitTask:
			// 收到退出信号，worker退出
			return
		}
	}
}

// doMapTask 执行Map任务
func doMapTask(task Task, mapf func(string, string) []KeyValue) {
	// 读取输入文件
	content, err := os.ReadFile(task.FileName)
	if err != nil {
		log.Printf("Cannot read file %v: %v", task.FileName, err)
		return
	}

	// 执行Map函数
	kva := mapf(task.FileName, string(content))

	// 创建临时文件来存储中间结果
	intermediate := make([][]KeyValue, task.NReduce)
	for i := range intermediate {
		intermediate[i] = []KeyValue{}
	}

	// 将结果分配到对应的reduce任务
	for _, kv := range kva {
		reduceTaskId := ihash(kv.Key) % task.NReduce
		intermediate[reduceTaskId] = append(intermediate[reduceTaskId], kv)
	}

	// 将中间结果写入文件
	for i := range intermediate {
		// 创建临时文件
		tempFile, err := os.CreateTemp("", fmt.Sprintf("mr-%d-%d-*", task.Id, i))
		if err != nil {
			log.Printf("Cannot create temp file: %v", err)
			return
		}

		// 将结果写入临时文件
		enc := json.NewEncoder(tempFile)
		for _, kv := range intermediate[i] {
			err := enc.Encode(&kv)
			if err != nil {
				log.Printf("Cannot encode key-value pair: %v", err)
				tempFile.Close()
				os.Remove(tempFile.Name())
				return
			}
		}
		tempFile.Close()

		// 原子性地重命名临时文件
		oname := fmt.Sprintf("mr-%d-%d", task.Id, i)
		err = os.Rename(tempFile.Name(), oname)
		if err != nil {
			log.Printf("Cannot rename temp file: %v", err)
			os.Remove(tempFile.Name())
			return
		}
	}

	// 报告任务完成
	reportTask(task.Id, MapTask)
}

// doReduceTask 执行Reduce任务
func doReduceTask(task Task, reducef func(string, []string) string) {
	// 收集所有属于这个reduce任务的中间文件
	kva := []KeyValue{}
	for i := 0; i < task.NMap; i++ {
		filename := fmt.Sprintf("mr-%d-%d", i, task.Id)
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Cannot open file %v: %v", filename, err)
			continue
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}

	// 按键排序
	sort.Slice(kva, func(i, j int) bool {
		return kva[i].Key < kva[j].Key
	})

	// 创建临时输出文件
	tempFile, err := os.CreateTemp("", fmt.Sprintf("mr-out-%d-*", task.Id))
	if err != nil {
		log.Printf("Cannot create temp file: %v", err)
		return
	}

	// 执行Reduce
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)

		// 将结果写入文件
		fmt.Fprintf(tempFile, "%v %v\n", kva[i].Key, output)
		i = j
	}
	tempFile.Close()

	// 原子性地重命名输出文件
	oname := fmt.Sprintf("mr-out-%d", task.Id)
	err = os.Rename(tempFile.Name(), oname)
	if err != nil {
		log.Printf("Cannot rename temp file: %v", err)
		os.Remove(tempFile.Name())
		return
	}

	// 清理中间文件
	for i := 0; i < task.NMap; i++ {
		filename := fmt.Sprintf("mr-%d-%d", i, task.Id)
		os.Remove(filename)
	}

	// 报告任务完成
	reportTask(task.Id, ReduceTask)
}

// reportTask 向coordinator报告任务完成
func reportTask(taskId int, taskType TaskType) {
	args := ReportTaskArgs{
		TaskId:   taskId,
		TaskType: taskType,
	}
	reply := ReportTaskReply{}
	call("Coordinator.ReportTask", &args, &reply)
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
