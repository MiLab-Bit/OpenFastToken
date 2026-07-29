package workerpool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubmitNilPoolRuns(t *testing.T) {
	globalPool = nil
	done := make(chan int, 1)
	Submit(func(ctx context.Context) { done <- 1 })
	select {
	case v := <-done:
		assert.Equal(t, 1, v)
	case <-time.After(2 * time.Second):
		t.Fatal("task not executed")
	}
}

func newPool(workers, q int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{tasks: make(chan Task, q), workers: workers, ctx: ctx, cancel: cancel}
	p.start()
	return p
}

func TestSubmitViaPool(t *testing.T) {
	globalPool = newPool(4, 100)
	defer func() {
		if globalPool != nil {
			globalPool.cancel()
			close(globalPool.tasks)
		}
	}()
	done := make(chan int, 1)
	Submit(func(ctx context.Context) { done <- 7 })
	select {
	case v := <-done:
		assert.Equal(t, 7, v)
	case <-time.After(2 * time.Second):
		t.Fatal("task not executed via pool")
	}
}

func TestSubmitQueueFullDrops(t *testing.T) {
	globalPool = &Pool{tasks: make(chan Task, 1)}
	globalPool.ctx, globalPool.cancel = context.WithCancel(context.Background())
	// fill the single slot
	globalPool.tasks <- func(ctx context.Context) {}
	// Submit should hit the default (drop) branch and return immediately (non-blocking)
	done := make(chan struct{})
	go func() {
		Submit(func(ctx context.Context) {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Submit blocked on full queue")
	}
}
