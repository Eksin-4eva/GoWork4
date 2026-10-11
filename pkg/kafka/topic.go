package kafka

import (
	"context"
	"fmt"
	"net"
	"strconv"

	kafkago "github.com/segmentio/kafka-go"
)

// dialNetwork 是 Kafka 的传输层协议。
const dialNetwork = "tcp"

// EnsureTopics 确保这些 topic 已存在，不存在就创建。
//
// 为什么不能指望 broker 的自动建 topic：
//   - 生产集群通常会显式关掉自动创建，靠它会在上线时翻车
//   - 即使开着，首次生产请求也会和建 topic 竞争，直接报
//     UnknownTopicOrPartition 把第一条事件丢掉
//
// 这个错误是真踩到的：投稿事件一条都没发出去，worker 什么都没收到。
func (b *Broker) EnsureTopics(ctx context.Context, topics []string, partitions int) error {
	if len(b.Producer.brokers) == 0 {
		return fmt.Errorf("kafka.EnsureTopics: brokers 未配置")
	}

	conn, err := kafkago.DialContext(ctx, dialNetwork, b.Producer.brokers[0])
	if err != nil {
		return fmt.Errorf("kafka.EnsureTopics: dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	existing, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("kafka.EnsureTopics: read partitions: %w", err)
	}

	seen := make(map[string]struct{}, len(existing))
	for i := range existing {
		seen[existing[i].Topic] = struct{}{}
	}

	missing := make([]kafkago.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		if _, ok := seen[topic]; ok {
			continue
		}
		missing = append(missing, kafkago.TopicConfig{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		})
	}
	if len(missing) == 0 {
		return nil
	}

	// 建 topic 必须找 controller，其他节点会拒绝
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("kafka.EnsureTopics: find controller: %w", err)
	}
	controllerConn, err := kafkago.DialContext(ctx, dialNetwork,
		net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("kafka.EnsureTopics: dial controller: %w", err)
	}
	defer func() { _ = controllerConn.Close() }()

	if err := controllerConn.CreateTopics(missing...); err != nil {
		return fmt.Errorf("kafka.EnsureTopics: create: %w", err)
	}
	return nil
}
