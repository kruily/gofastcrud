package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kruily/gofastcrud/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Job 任务接口
type Job interface {
	Run(ctx context.Context) error
	GetName() string
}

// JobStatus 任务状态
type JobStatus struct {
	LastRun   time.Time
	NextRun   time.Time
	LastError error
	Status    string
	EntryID   cron.EntryID
}

// Scheduler cron调度器
type Scheduler struct {
	cron   *cron.Cron
	jobs   map[string]Job
	status map[string]JobStatus
	mu     sync.RWMutex
	logger *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// Options 调度器配置
type Options struct {
	Logger   *logger.Logger
	Location *time.Location
}

// NewScheduler 创建调度器实例
func NewScheduler(ctx context.Context, opt Options) *Scheduler {
	if opt.Location == nil {
		opt.Location = time.Local
	}

	ctx, cancel := context.WithCancel(ctx)

	if opt.Logger == nil {
		logConfig := logger.Config{
			Level: logger.InfoLevel,
			FileConfig: &logger.FileConfig{
				Filename:   "logs/scheduler.log",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   true,
			},
			ConsoleLevel: logger.DebugLevel,
		}
		defaultLogger, err := logger.NewLogger(logConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to create default logger: %v", err))
		}
		opt.Logger = defaultLogger
	}

	return &Scheduler{
		cron:   cron.New(cron.WithSeconds(), cron.WithLocation(opt.Location)),
		jobs:   make(map[string]Job),
		status: make(map[string]JobStatus),
		logger: opt.Logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddJob 添加定时任务
func (s *Scheduler) AddJob(spec string, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.GetName()]; exists {
		s.logger.Warn("Job already exists", map[string]interface{}{
			"job": job.GetName(),
		})
		return fmt.Errorf("job %s already exists", job.GetName())
	}

	entryID, err := s.cron.AddFunc(spec, func() {
		if err := job.Run(s.ctx); err != nil {
			s.mu.Lock()
			s.status[job.GetName()] = JobStatus{
				LastRun:   time.Now(),
				LastError: err,
				Status:    "error",
			}
			s.mu.Unlock()
			s.logger.Error("Job failed", map[string]interface{}{
				"job":   job.GetName(),
				"error": err.Error(),
			})
		} else {
			s.mu.Lock()
			s.status[job.GetName()] = JobStatus{
				LastRun: time.Now(),
				Status:  "completed",
			}
			s.mu.Unlock()
			s.logger.Info("Job completed", map[string]interface{}{
				"job": job.GetName(),
			})
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add job: %v", err)
	}

	s.jobs[job.GetName()] = job
	s.status[job.GetName()] = JobStatus{
		Status:  "scheduled",
		EntryID: entryID,
	}

	s.logger.Info("Job added", map[string]interface{}{
		"job":  job.GetName(),
		"spec": spec,
	})
	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.logger.Info("Starting scheduler", nil)
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler", nil)
	s.cancel()
	s.cron.Stop()
}

// Remove 移除任务
func (s *Scheduler) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.status[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	s.cron.Remove(status.EntryID)
	delete(s.jobs, name)
	delete(s.status, name)

	s.logger.Info("Job removed", map[string]interface{}{
		"job": name,
	})
	return nil
}

// GetStatus 获取任务状态
func (s *Scheduler) GetStatus(name string) (JobStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.status[name]
	if !exists {
		return JobStatus{}, fmt.Errorf("job %s not found", name)
	}

	return status, nil
}
