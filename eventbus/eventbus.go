package eventbus

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// Event 事件接口
type Event interface {
	// EventName 返回事件名称
	EventName() string
}

// Handler 事件处理器函数类型
type EventHandler func(ctx context.Context, event Event) error

// EventBus 事件总线
type EventBus struct {
	handlers     map[string][]EventHandler
	asyncHandler chan eventWrapper
	mu           sync.RWMutex
	wg           sync.WaitGroup
	workerCount  int
	timeout      time.Duration
}

// eventWrapper 事件包装器
type eventWrapper struct {
	ctx     context.Context
	event   Event
	handler EventHandler
}

// Options 事件总线配置选项
type Options struct {
	WorkerCount int           // 异步处理的工作协程数
	QueueSize   int           // 异步事件队列大小
	Timeout     time.Duration // 事件处理超时时间
}

// DefaultOptions 默认配置
var DefaultOptions = Options{
	WorkerCount: 10,
	QueueSize:   1000,
	Timeout:     time.Second * 30,
}

// New 创建事件总线
func New(opts ...Options) *EventBus {
	opt := DefaultOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	bus := &EventBus{
		handlers:     make(map[string][]EventHandler),
		asyncHandler: make(chan eventWrapper, opt.QueueSize),
		workerCount:  opt.WorkerCount,
		timeout:      opt.Timeout,
	}

	// 启动工作协程
	for i := 0; i < opt.WorkerCount; i++ {
		go bus.worker()
	}

	return bus
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.handlers[eventName]; !ok {
		b.handlers[eventName] = make([]EventHandler, 0)
	}
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// SubscribeAsync 异步订阅事件
func (b *EventBus) SubscribeAsync(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.handlers[eventName]; !ok {
		b.handlers[eventName] = make([]EventHandler, 0)
	}

	// 直接添加异步处理器
	asyncHandler := func(ctx context.Context, event Event) error {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			timeoutCtx, cancel := context.WithTimeout(ctx, b.timeout)
			defer cancel()

			handler(timeoutCtx, event)
		}()
		return nil
	}

	b.handlers[eventName] = append(b.handlers[eventName], asyncHandler)
}

// Publish 发布事件（同步）
func (b *EventBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers, exists := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if !exists {
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// PublishAsync 异步发布事件
func (b *EventBus) PublishAsync(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers, exists := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		b.wg.Add(1)
		go func(h EventHandler) {
			defer b.wg.Done()
			timeoutCtx, cancel := context.WithTimeout(ctx, b.timeout)
			defer cancel()

			h(timeoutCtx, event)
		}(handler)
	}
}

// Wait 等待所有异步事件处理完成
func (b *EventBus) Wait() {
	b.wg.Wait()
}

// worker 工作协程
func (b *EventBus) worker() {
	for wrapper := range b.asyncHandler {
		func() {
			timeoutCtx, cancel := context.WithTimeout(wrapper.ctx, b.timeout)
			defer cancel()

			wrapper.handler(timeoutCtx, wrapper.event)
		}()
	}
}

// HasSubscriber 检查是否有订阅者
func (b *EventBus) HasSubscriber(eventName string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, exists := b.handlers[eventName]
	return exists && len(handlers) > 0
}

// Unsubscribe 取消订阅
func (b *EventBus) Unsubscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, exists := b.handlers[eventName]; exists {
		handlerPtr := reflect.ValueOf(handler).Pointer()
		newHandlers := make([]EventHandler, 0)

		for _, h := range handlers {
			if reflect.ValueOf(h).Pointer() != handlerPtr {
				newHandlers = append(newHandlers, h)
			}
		}

		b.handlers[eventName] = newHandlers
	}
}

// Close 关闭事件总线
func (b *EventBus) Close() {
	close(b.asyncHandler)
	b.Wait()
}
