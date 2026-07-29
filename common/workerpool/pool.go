package workerpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/MiLab-Bit/OpenFastToken/logger"
)

// Task 异步任务函数类型
type Task func(ctx context.Context)

var (
	globalPool *Pool
	once       sync.Once
)

// Pool 轻量级 Worker Pool
type Pool struct {
	tasks   chan Task
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// Init 初始化全局 Worker Pool
func Init(workers, queueSize int) {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		globalPool = &Pool{
			tasks:   make(chan Task, queueSize),
			workers: workers,
			ctx:     ctx,
			cancel:  cancel,
		}
		globalPool.start()
	})
}

// Submit 提交异步任务（非阻塞，队列满时丢弃并打印警告）
func Submit(task Task) {
	if globalPool == nil {
		// 未初始化，同步执行
		go task(context.Background())
		return
	}
	select {
	case globalPool.tasks <- task:
		// 提交成功
	default:
		logger.LogWarn(globalPool.ctx, "WorkerPool queue is full, dropping async task")
	}
}

// Shutdown 优雅关闭，等待所有任务完成
func Shutdown() {
	if globalPool != nil {
		globalPool.cancel()
		close(globalPool.tasks)
		globalPool.wg.Wait()
	}
}

func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogError(p.ctx, fmt.Sprintf("WorkerPool task panic: %v", r))
				}
			}()
			task(p.ctx)
		}()
	}
}
