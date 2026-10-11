package kafka

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// 读取参数。
const (
	readMinBytes = 1
	readMaxBytes = 10 << 20 // 10MB，够放最大的事件体
	// errorBackoff 是读取出错后的退避时间，避免出错时把 CPU 打满。
	errorBackoff = time.Second
)

// Consumer 按 topic + 消费组读取事件。
type Consumer struct {
	brokers []string
	groupID string

	mu      sync.Mutex
	readers []*kafkago.Reader
}

// NewConsumer 构造消费者。
func NewConsumer(brokers []string, groupID string) *Consumer {
	return &Consumer{brokers: brokers, groupID: groupID}
}

// Start 启动 num 个读取协程，把消息送到返回的 channel。
//
// 调用方**不要自己关这个 channel**：它由本包在所有读取协程退出后关闭，
// 提前关会让正在发送的协程 panic。
func (c *Consumer) Start(ctx context.Context, topic string, num int) <-chan *Message {
	if num <= 0 {
		num = 1
	}

	ch := make(chan *Message)
	var wg sync.WaitGroup

	for i := 0; i < num; i++ {
		reader := c.newReader(topic)
		c.track(reader)

		wg.Add(1)
		go func(r *kafkago.Reader) {
			defer wg.Done()
			c.pump(ctx, r, ch)
		}(reader)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// Close 关闭所有 reader。
func (c *Consumer) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.readers = nil
	return firstErr
}

// pump 从单个 reader 读消息并转发到 channel，直到 ctx 结束。
func (c *Consumer) pump(ctx context.Context, reader *kafkago.Reader, ch chan<- *Message) {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			// ctx 取消与 reader 关闭都属于正常退出
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) || ctx.Err() != nil {
				return
			}
			// 其他错误（比如 broker 暂时不可达）退避后重试
			select {
			case <-time.After(errorBackoff):
			case <-ctx.Done():
				return
			}
			continue
		}

		select {
		case ch <- &Message{Key: msg.Key, Value: msg.Value}:
		case <-ctx.Done():
			return
		}
	}
}

// newReader 创建一个消费组 reader。
//
// 不显式关闭自动提交（CommitInterval 保持默认 1 秒），
// 因此投递语义是 at-least-once：崩溃前最多 1 秒的 offset 可能没提交，
// 重启后这些消息会重投。**这是刻意的选择**——消费端全部做成幂等的，
// 比自己去管 offset 提交简单得多，也更不容易出错。
func (c *Consumer) newReader(topic string) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  c.groupID,
		Topic:    topic,
		MinBytes: readMinBytes,
		MaxBytes: readMaxBytes,
	})
}

func (c *Consumer) track(reader *kafkago.Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readers = append(c.readers, reader)
}
