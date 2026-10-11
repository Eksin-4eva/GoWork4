// Package event 定义跨进程传递的事件体。
//
// 生产者（api）与消费者（worker）必须对同一份结构达成一致，
// 所以单独放一个包、两边共同引用，避免各写一份结构体然后悄悄漂移。
package event

// VideoUploaded 是投稿成功事件。
type VideoUploaded struct {
	V        int    `json:"v"`
	VideoID  int64  `json:"video_id"`
	VideoKey string `json:"video_key"`
}

// VideoVisited 是播放事件。
//
// EventID 用于消费端去重：计数类操作天然不幂等（+1 执行两次就是多算一次），
// 而 Kafka 是 at-least-once，必须能识别出「这条我处理过了」。
type VideoVisited struct {
	V       int    `json:"v"`
	EventID string `json:"event_id"`
	VideoID int64  `json:"video_id"`
}
