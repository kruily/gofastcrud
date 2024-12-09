package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Job 任务接口
type Job interface {
	// Run 执行任务
	Run(ctx context.Context) error
	// GetName 获取任务名称
	GetName() string
}

// JobStatus 任务状态
type JobStatus struct {
	Name       string    // 任务名称
	LastRun    time.Time // 上次运行时间
	NextRun    time.Time // 下次运行时间
	LastError  error     // 上次错误
	RunCount   int64     // 运行次数
	Status     string    // 状态：running, waiting, stopped
	CreateTime time.Time // 创建时间
}

// Scheduler 调度器
type Scheduler struct {
	cron    *cron.Cron           // cron调度器
	jobs    map[string]Job       // 所有任务
	status  map[string]JobStatus // 任务状态
	mu      sync.RWMutex
	logger  Logger
	wg      sync.WaitGroup
	oneTime map[string]*time.Timer // 一次性任务
	ctx     context.Context
	cancel  context.CancelFunc
}

// Options 调度器配置
type Options struct {
	Logger     Logger         // 日志记录器
	Location   *time.Location // 时区
	MaxRetries int            // 最大重试次数
}

// New 创建调度器
func New(opts ...Options) *Scheduler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.Logger == nil {
		opt.Logger = NewDefaultLogger()
	}

	if opt.Location == nil {
		opt.Location = time.Local
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:    cron.New(cron.WithLocation(opt.Location)),
		jobs:    make(map[string]Job),
		status:  make(map[string]JobStatus),
		logger:  opt.Logger,
		oneTime: make(map[string]*time.Timer),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// AddCronJob 添加定时任务
func (s *Scheduler) AddCronJob(spec string, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.GetName()]; exists {
		return fmt.Errorf("job %s already exists", job.GetName())
	}

	_, err := s.cron.AddFunc(spec, func() {
		s.runJob(job)
	})
	if err != nil {
		return err
	}

	s.jobs[job.GetName()] = job
	s.status[job.GetName()] = JobStatus{
		Name:       job.GetName(),
		Status:     "waiting",
		CreateTime: time.Now(),
	}

	return nil
}

// AddOneTimeJob 添加一次性任务
func (s *Scheduler) AddOneTimeJob(delay time.Duration, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.GetName()]; exists {
		return fmt.Errorf("job %s already exists", job.GetName())
	}

	timer := time.AfterFunc(delay, func() {
		s.runJob(job)
		s.removeJob(job.GetName())
	})

	s.oneTime[job.GetName()] = timer
	s.jobs[job.GetName()] = job
	s.status[job.GetName()] = JobStatus{
		Name:       job.GetName(),
		Status:     "waiting",
		CreateTime: time.Now(),
		NextRun:    time.Now().Add(delay),
	}

	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
	s.cron.Stop()

	// 停止所有一次性任务
	s.mu.Lock()
	for _, timer := range s.oneTime {
		timer.Stop()
	}
	s.mu.Unlock()

	// 等待所有任务完成
	s.wg.Wait()
}

// GetJobStatus 获取任务状态
func (s *Scheduler) GetJobStatus(name string) (JobStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.status[name]
	return status, exists
}

// ListJobs 列出所有任务
func (s *Scheduler) ListJobs() []JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]JobStatus, 0, len(s.status))
	for _, status := range s.status {
		jobs = append(jobs, status)
	}
	return jobs
}

// RemoveJob 移除任务
func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.removeJob(name)
}

// removeJob 内部移除任务方法
func (s *Scheduler) removeJob(name string) error {
	if timer, exists := s.oneTime[name]; exists {
		timer.Stop()
		delete(s.oneTime, name)
	}

	delete(s.jobs, name)
	delete(s.status, name)
	return nil
}

// runJob 执行任务
func (s *Scheduler) runJob(job Job) {
	s.wg.Add(1)
	defer s.wg.Done()

	s.mu.Lock()
	s.status[job.GetName()] = JobStatus{
		Name:       job.GetName(),
		LastRun:    time.Now(),
		Status:     "running",
		RunCount:   s.status[job.GetName()].RunCount + 1,
		CreateTime: s.status[job.GetName()].CreateTime,
	}
	s.mu.Unlock()

	// 执行任务
	err := job.Run(s.ctx)

	s.mu.Lock()
	status := s.status[job.GetName()]
	status.LastError = err
	status.Status = "waiting"
	s.status[job.GetName()] = status
	s.mu.Unlock()

	if err != nil {
		s.logger.Error("job failed",
			map[string]interface{}{
				"job":   job.GetName(),
				"error": err.Error(),
			})
	}
}
