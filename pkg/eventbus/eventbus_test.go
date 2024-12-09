package eventbus

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

// TestEvent 测试用的事件
type TestEvent struct {
	name    string
	payload string
}

func (e TestEvent) EventName() string {
	return e.name
}

func TestEventBus(t *testing.T) {
	bus := New()
	defer bus.Close()

	// 测试同步订阅和发布
	t.Run("SyncSubscribeAndPublish", func(t *testing.T) {
		var count int32
		eventName := "test_event"

		// 订阅事件
		bus.Subscribe(eventName, func(ctx context.Context, event Event) error {
			log.Default().Println("test_event")
			atomic.AddInt32(&count, 1)
			return nil
		})
		log.Default().Println("test_event_publish")
		// 发布事件
		event := TestEvent{name: eventName, payload: "test"}
		if err := bus.Publish(context.Background(), event); err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}

		if atomic.LoadInt32(&count) != 1 {
			t.Errorf("Expected count to be 1, got %d", count)
		}
	})

	// 测试异步订阅和发布
	t.Run("AsyncSubscribeAndPublish", func(t *testing.T) {
		var count int32
		eventName := "test_async_event"

		// 异步订阅事件
		bus.SubscribeAsync(eventName, func(ctx context.Context, event Event) error {
			time.Sleep(100 * time.Millisecond) // 模拟异步处理
			atomic.AddInt32(&count, 1)
			return nil
		})

		// 异步发布多个事件
		for i := 0; i < 5; i++ {
			event := TestEvent{name: eventName, payload: "test"}
			bus.PublishAsync(context.Background(), event)
		}

		// 等待所有事件处理完成
		bus.Wait()

		if atomic.LoadInt32(&count) != 5 {
			t.Errorf("Expected count to be 5, got %d", count)
		}
	})

	// 测试错误处理
	t.Run("ErrorHandling", func(t *testing.T) {
		eventName := "error_event"
		expectedError := errors.New("test error")

		// 订阅返回错误的处理器
		bus.Subscribe(eventName, func(ctx context.Context, event Event) error {
			return expectedError
		})

		// 发布事件并检查错误
		event := TestEvent{name: eventName, payload: "test"}
		if err := bus.Publish(context.Background(), event); err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
	})

	// 测试超时处理
	t.Run("Timeout", func(t *testing.T) {
		bus := New(Options{
			WorkerCount: 1,
			QueueSize:   1,
			Timeout:     100 * time.Millisecond,
		})
		defer bus.Close()

		eventName := "timeout_event"
		var completed int32

		// 订阅一个耗时的处理器
		bus.Subscribe(eventName, func(ctx context.Context, event Event) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				atomic.AddInt32(&completed, 1)
				return nil
			}
		})

		// 发布事件
		event := TestEvent{name: eventName, payload: "test"}
		bus.PublishAsync(context.Background(), event)
		bus.Wait()

		// 由于超时，处理器应该没有完成
		if atomic.LoadInt32(&completed) != 0 {
			t.Error("Handler should not have completed due to timeout")
		}
	})

	// 测试取消订阅
	t.Run("Unsubscribe", func(t *testing.T) {
		eventName := "unsubscribe_event"
		var count int32

		handler := func(ctx context.Context, event Event) error {
			atomic.AddInt32(&count, 1)
			return nil
		}

		// 订阅事件
		bus.Subscribe(eventName, handler)

		// 发布第一个事件
		event := TestEvent{name: eventName, payload: "test"}
		bus.Publish(context.Background(), event)

		if atomic.LoadInt32(&count) != 1 {
			t.Error("Handler should have been called once")
		}

		// 取消订阅
		bus.Unsubscribe(eventName, handler)

		// 发布第二个事件
		bus.Publish(context.Background(), event)

		if atomic.LoadInt32(&count) != 1 {
			t.Error("Handler should not have been called after unsubscribe")
		}
	})

	// 测试并发订阅和发布
	t.Run("ConcurrentSubscribeAndPublish", func(t *testing.T) {
		const numEvents = 100
		const numHandlers = 5
		var totalCount int32

		eventName := "concurrent_event"

		// 添加多个处理器
		for i := 0; i < numHandlers; i++ {
			bus.Subscribe(eventName, func(ctx context.Context, event Event) error {
				atomic.AddInt32(&totalCount, 1)
				return nil
			})
		}

		// 并发发布事件
		done := make(chan bool)
		for i := 0; i < numEvents; i++ {
			go func() {
				event := TestEvent{name: eventName, payload: "test"}
				bus.Publish(context.Background(), event)
				done <- true
			}()
		}

		// 等待所有事件发布完成
		for i := 0; i < numEvents; i++ {
			<-done
		}

		expectedCount := int32(numEvents * numHandlers)
		if atomic.LoadInt32(&totalCount) != expectedCount {
			t.Errorf("Expected count to be %d, got %d", expectedCount, totalCount)
		}
	})
}
