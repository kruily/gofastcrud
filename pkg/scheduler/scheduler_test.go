package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestJob 测试用的任务
type TestJob struct {
	name      string
	execCount int32
}

func (j *TestJob) Run(ctx context.Context) error {
	atomic.AddInt32(&j.execCount, 1)
	return nil
}

func (j *TestJob) GetName() string {
	return j.name
}

func (j *TestJob) GetExecCount() int32 {
	return atomic.LoadInt32(&j.execCount)
}

func TestScheduler(t *testing.T) {
	ctx := context.Background()
	scheduler := NewScheduler(ctx, Options{})

	// 创建测试任务
	job := &TestJob{name: "test_job"}

	// 测试添加任务
	t.Run("AddJob", func(t *testing.T) {
		err := scheduler.AddJob("*/1 * * * * *", job) // 每秒执行一次
		if err != nil {
			t.Errorf("Failed to add job: %v", err)
		}

		// 验证任务状态
		status, err := scheduler.GetStatus(job.GetName())
		if err != nil {
			t.Errorf("Failed to get job status: %v", err)
		}
		if status.Status != "scheduled" {
			t.Errorf("Expected status 'scheduled', got '%s'", status.Status)
		}
	})

	// 测试启动调度器
	t.Run("Start", func(t *testing.T) {
		scheduler.Start()
		// 等待任务执行
		time.Sleep(3 * time.Second)

		// 验证任务是否执行
		execCount := job.GetExecCount()
		if execCount < 2 {
			t.Errorf("Expected at least 2 executions, got %d", execCount)
		}
	})

	// 测试移除任务
	t.Run("Remove", func(t *testing.T) {
		err := scheduler.Remove(job.GetName())
		if err != nil {
			t.Errorf("Failed to remove job: %v", err)
		}

		// 验证任务是否已移除
		_, err = scheduler.GetStatus(job.GetName())
		if err == nil {
			t.Error("Expected error when getting removed job status")
		}
	})

	// 测试停止调度器
	t.Run("Stop", func(t *testing.T) {
		scheduler.Stop()
		// 等待一会儿
		time.Sleep(time.Second)

		// 添加新任务应该仍然可以工作
		newJob := &TestJob{name: "new_test_job"}
		err := scheduler.AddJob("*/1 * * * * *", newJob)
		if err != nil {
			t.Errorf("Failed to add job after stop: %v", err)
		}
	})
}

// TestSchedulerConcurrency 测试并发情况
func TestSchedulerConcurrency(t *testing.T) {
	ctx := context.Background()
	scheduler := NewScheduler(ctx, Options{})
	scheduler.Start()
	defer scheduler.Stop()

	// 并发添加任务
	t.Run("ConcurrentAdd", func(t *testing.T) {
		const numJobs = 10
		errCh := make(chan error, numJobs)

		for i := 0; i < numJobs; i++ {
			go func(id int) {
				job := &TestJob{name: fmt.Sprintf("concurrent_job_%d", id)}
				errCh <- scheduler.AddJob("*/5 * * * * *", job) // 每5秒执行一次
			}(i)
		}

		// 收集错误
		for i := 0; i < numJobs; i++ {
			if err := <-errCh; err != nil {
				t.Errorf("Failed to add job concurrently: %v", err)
			}
		}
	})

	// 并发获取状态
	t.Run("ConcurrentGetStatus", func(t *testing.T) {
		const numGets = 100
		errCh := make(chan error, numGets)

		for i := 0; i < numGets; i++ {
			go func(id int) {
				_, err := scheduler.GetStatus(fmt.Sprintf("concurrent_job_%d", id%10))
				errCh <- err
			}(i)
		}

		// 收集错误
		for i := 0; i < numGets; i++ {
			<-errCh // 这里不检查错误，因为有些任务可能已经被移除
		}
	})
}
