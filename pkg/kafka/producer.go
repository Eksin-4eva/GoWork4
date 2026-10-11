package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// writeTimeout 是单次写入的超时。
//
// 必须有：Kafka 不可用时，写操作会一直卡在重试上，
// 把调用它的 HTTP 请求线程一起拖死。
const writeTimeout = 5 * time.Second

// Producer 按 topic 复用 Writer。
type Producer struct {
	brokers []string

	mu      sync.Mutex
	writers map[string]*kafkago.Writer
}

// NewProducer 构造生产者。
func NewProducer(brokers []string) *Producer {
	return &Producer{brokers: brokers, writers: make(map[string]*kafkago.Writer)}
}

// Send 把 value 序列化成 JSON 发到指定 topic。
//
// key 决定分区：同一个 key 的消息落到同一分区，从而保持顺序。
// 留空会让消息在各分区之间轮询，同一实体的事件就可能乱序。
func (p *Producer) Send(ctx context.Context, topic string, key []byte, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("kafka.Producer.Send: marshal %s: %w", topic, err)
	}

	writer, err := p.writerFor(topic)
	if err != nil {
		return err
	}

	if err := writer.WriteMessages(ctx, kafkago.Message{Key: key, Value: payload}); err != nil {
		return fmt.Errorf("kafka.Producer.Send: write %s: %w", topic, err)
	}
	return nil
}

// Close 关闭所有 writer。
func (p *Producer) Close() error {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("kafka.Producer.Close: %s: %w", topic, err)
		}
	}
	p.writers = make(map[string]*kafkago.Writer)
	return firstErr
}

// writerFor 返回（必要时创建）某个 topic 的 writer。
func (p *Producer) writerFor(topic string) (*kafkago.Writer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if writer, ok := p.writers[topic]; ok {
		return writer, nil
	}
	if len(p.brokers) == 0 {
		return nil, fmt.Errorf("kafka.Producer: brokers 未配置")
	}

	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(p.brokers...),
		Topic:    topic,
		Balancer: &kafkago.Hash{}, // 按 key 分区，保证同一实体的事件有序
		// 只要 leader 确认即可：失败有补偿，但主链路不能被写延迟拖住
		RequiredAcks: kafkago.RequireOne,
		WriteTimeout: writeTimeout,
	}
	p.writers[topic] = writer
	return writer, nil
}
