// Package kafka 封装事件的生产与消费。
//
// 与 pkg/db 的错误策略不同：这里的失败是**可降级**的。
// 消息队列不可用时主链路必须照常返回，所以本包只返回带上下文的原始错误，
// 由调用方决定是降级、丢弃还是上报。
package kafka

// Message 是一条消费到的事件。
type Message struct {
	Key   []byte
	Value []byte
}

// Broker 聚合生产者与消费者，供 ClientSet 注入。
type Broker struct {
	Producer *Producer
	Consumer *Consumer
}

// NewBroker 构造消息队列封装。
//
// 此时不会建立任何连接——kafka-go 是首次读写时才拨号的，
// 所以 Kafka 没起也不影响服务启动。
func NewBroker(brokers []string, groupID string) *Broker {
	return &Broker{
		Producer: NewProducer(brokers),
		Consumer: NewConsumer(brokers, groupID),
	}
}

// Close 释放底层连接。
func (b *Broker) Close() error {
	if b == nil {
		return nil
	}

	var firstErr error
	if b.Producer != nil {
		firstErr = b.Producer.Close()
	}
	if b.Consumer != nil {
		if err := b.Consumer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
