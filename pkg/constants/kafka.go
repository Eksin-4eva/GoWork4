package constants

import "time"

// 事件 topic。命名规则 gobili.<域>.<事件>。
const (
	// TopicVideoUploaded 在投稿成功后发出，worker 消费后抽帧生成封面。
	TopicVideoUploaded = "gobili.video.uploaded"
	// TopicVideoVisited 在每次播放时发出，worker 聚合后批量写库。
	TopicVideoVisited = "gobili.video.visited"
)

// 事件体与消费相关参数。
const (
	// SchemaVersion 是事件体的版本号。
	// 消费者按版本兼容处理，避免「加一个字段要同时升级生产端和消费端」。
	SchemaVersion = 1

	// TopicPartitions 是新建 topic 的分区数。
	//
	// 分区数决定消费并发上限：同一个 key 的消息只落一个分区，
	// 分区太少会让单个热门视频的事件串行化。
	TopicPartitions = 3

	// 消费者协程数。抽帧是 CPU 密集操作，所以协程数少于计数聚合。
	UploadedConsumerNum = 2
	VisitedConsumerNum  = 4

	// 播放计数的批量聚合参数：攒够 Or 等够就提交一次。
	VisitFlushInterval = 2 * time.Second
	VisitFlushSize     = 200

	// DedupTTL 是事件去重集合的存活时间。
	// 只需覆盖重复投递可能出现的窗口，不需要永久保留。
	DedupTTL = 10 * time.Minute

	// CoverCompensateInterval 是封面补偿任务的扫描间隔。
	//
	// 封面事件一旦丢了，用户就永远看不到封面——这是唯一「丢了无法自愈」的链路，
	// 所以必须有兜底扫描。
	CoverCompensateInterval = time.Minute
	CoverCompensateBatch    = 50
)

// KafkaTopics 返回本项目用到的全部 topic。
func KafkaTopics() []string {
	return []string{TopicVideoUploaded, TopicVideoVisited}
}
