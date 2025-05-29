# Distributed MapReduce in Go

This repository contains my implementation of a distributed MapReduce framework in Go, based on MIT’s 6.5840 lab assignment. It demonstrates core distributed-systems concepts, robust engineering practices, and production-grade coding skills.

## 📦 Project Overview

I built two programs:

- **Coordinator** (master):  
  - Manages the global state of Map and Reduce tasks in memory  
  - Exposes two RPC endpoints:  
    1. `RequestTask` – “Give me work”  
    2. `ReportTask` – “I finished this task”  
  - Tracks task lifecycles (Idle → InProgress → Completed), enforces a 10 s timeout for fault tolerance, and transitions cleanly between Map and Reduce phases.

- **Worker**:  
  - Loops: ask for work → execute → report → repeat  
  - Implements `doMapTask` and `doReduceTask`:  
    - **Map**: reads an input split, applies `mapf`, partitions output via `ihash` into `mr-<mapID>-<reduceID>` files  
    - **Reduce**: reads all `mr-*-<reduceID>` files, sorts & groups by key, applies `reducef`, and writes final `mr-out-<reduceID>` atomically  
  - Handles “no work” (backs off with sleep) and “exit” signals gracefully.

## 🚀 Key Engineering Highlights

- **Concurrency & Synchronization**  
  - Coordinated via `sync.Mutex` to protect shared task state  
  - Per-RPC locking ensures correctness under high concurrency

- **Fault Tolerance & Retry**  
  - Unfinished tasks automatically re-queued after 10 s  
  - Graceful recovery from worker crashes (tested with random “crash” plugin)

- **Atomic File I/O**  
  - Uses Go’s `os.CreateTemp` + `os.Rename` to guarantee atomic writes  
  - Clean-up of stale or partial files on failures

- **Clean RPC Design**  
  - Minimal, well-typed Go `net/rpc` interfaces  
  - UNIX-domain sockets simplify local testing while mirroring real-world RPC

- **Testable & Race-Free**  
  - Comprehensive end-to-end test script (`test-mr.sh`) validates correctness, parallelism, and crash recovery  
  - Verified with Go’s race detector (`go run -race`)

## 📐 Architecture Diagram

![Architecture Diagram](media/lab1.png)

