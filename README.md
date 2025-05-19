# 6.5840 Lab 1: MapReduce

## 项目简介
这是一个基于 Go 语言实现的分布式 MapReduce 系统。本实验要求实现一个完整的 MapReduce 框架，包括：
- Coordinator（协调者）进程：负责任务分发和失败处理
- Worker（工作者）进程：执行具体的 Map 和 Reduce 任务

系统需要能够处理：
- 并行执行 Map 和 Reduce 任务
- 处理 Worker 崩溃的情况
- 确保任务正确完成和输出

## 环境要求
- Go 语言环境（需要预先安装）
- Git 版本控制系统

## 项目结构
```
.
├── src/
│   ├── main/           # 主程序入口
│   │   ├── mrcoordinator.go  # Coordinator 启动程序
│   │   ├── mrworker.go       # Worker 启动程序
│   │   └── mrsequential.go   # 顺序执行版本（参考实现）
│   ├── mr/             # MapReduce 核心实现
│   │   ├── coordinator.go    # Coordinator 实现
│   │   ├── worker.go         # Worker 实现
│   │   └── rpc.go            # RPC 相关定义
│   └── mrapps/         # MapReduce 应用示例
│       ├── wc.go       # 词频统计应用
│       └── indexer.go  # 文本索引应用
├── Makefile           # 构建脚本
└── test-mr.sh         # 测试脚本
```

## 快速开始

### 1. 获取代码
```bash
git clone git://g.csail.mit.edu/6.5840-golabs-2024 6.5840
cd 6.5840
```

### 2. 运行顺序版本（参考实现）
```bash
cd src/main
go build -buildmode=plugin ../mrapps/wc.go
rm mr-out*
go run mrsequential.go wc.so pg*.txt
more mr-out-0
```

### 3. 运行分布式版本
1. 编译 MapReduce 应用：
```bash
cd src/main
go build -buildmode=plugin ../mrapps/wc.go
```

2. 启动 Coordinator：
```bash
rm mr-out*
go run mrcoordinator.go pg-*.txt
```

3. 在一个或多个终端窗口启动 Worker：
```bash
go run mrworker.go wc.so
```

4. 检查输出：
```bash
cat mr-out-* | sort | more
```

### 4. 运行测试
```bash
cd src/main
bash test-mr.sh
```

## 实现要求

### 基本要求
1. Map 阶段需要将中间键值对分成 nReduce 个桶，对应 nReduce 个 reduce 任务
2. Worker 需要将第 X 个 reduce 任务的输出写入 mr-out-X 文件
3. 输出文件格式要求：每行使用 "%v %v" 格式输出键值对
4. Worker 需要将中间 Map 输出保存在当前目录，以便后续 Reduce 任务读取
5. Coordinator 需要实现 Done() 方法，在 MapReduce 作业完成时返回 true
6. 作业完成时，Worker 进程需要正确退出

### 容错要求
1. Coordinator 需要处理 Worker 崩溃的情况
2. 如果 Worker 在 10 秒内未完成任务，Coordinator 需要将任务重新分配给其他 Worker
3. 需要确保在 Worker 崩溃的情况下，部分写入的文件不会影响最终结果

## 开发提示
1. 可以使用 Go 的 encoding/json 包处理中间文件的读写
2. 使用 ihash(key) 函数为键选择对应的 reduce 任务
3. 可以使用 ioutil.TempFile 和 os.Rename 确保文件写入的原子性
4. 使用 Go 的 race detector 检查并发问题：`go run -race`
5. 注意 RPC 调用时结构体字段需要首字母大写

## 可选挑战
1. 实现自己的 MapReduce 应用（如分布式 Grep）
2. 将 Coordinator 和 Worker 部署到不同机器上运行
3. 实现备份任务（Backup Tasks）机制

## 注意事项
1. 不要修改 main/mrcoordinator.go 和 main/mrworker.go
2. 可以修改 mr/coordinator.go、mr/worker.go 和 mr/rpc.go
3. 测试时会使用原始版本的其他文件
4. 确保代码能够通过所有测试用例
5. 注意处理并发情况下的数据竞争问题

## 测试说明
- test-mr.sh：运行基本功能测试
- test-mr-many.sh：多次运行测试以发现低概率 bug
- 测试包括：词频统计、文本索引、并行性、崩溃恢复等

## 提交说明
请确保代码能够通过所有测试用例，并遵循代码规范。提交前请运行完整的测试套件确保功能正常。 