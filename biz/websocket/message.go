package websocket

// Message 定义WebSocket消息结构
type Message struct {
	Type      int         `json:"type"`       // 消息类型：1-私聊，2-获取私聊历史，3-群聊，4-获取群聊历史
	FromUserID int64       `json:"from_user_id"` // 发送者ID
	ToUserID   int64       `json:"to_user_id,omitempty"`   // 接收者ID（私聊）
	RoomID     string      `json:"room_id,omitempty"`     // 聊天室ID（群聊）
	Content    string      `json:"content,omitempty"`    // 消息内容
	Page       int         `json:"page,omitempty"`       // 分页页码
	PageSize   int         `json:"page_size,omitempty"`  // 分页大小
	Timestamp  int64       `json:"timestamp"`  // 消息时间戳
	Data       interface{} `json:"data,omitempty"`       // 响应数据
}

// ChatHistory 定义聊天历史记录结构
type ChatHistory struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
}
