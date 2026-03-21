package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

var (
	upgrader = websocket.HertzUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(ctx *app.RequestContext) bool {
			return true // 允许所有来源的WebSocket连接
		},
	}

	manager = NewManager()
)

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(c context.Context, ctx *app.RequestContext) {
	// 从请求中获取用户ID（这里假设从JWT或其他方式获取）
	userIDStr := ctx.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Invalid user_id")
		return
	}

	// 升级HTTP连接为WebSocket连接
	err = upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		// 注册客户端连接
		manager.RegisterClient(userID, conn)
		log.Printf("User %d connected", userID)

		// 处理连接关闭
		defer func() {
			manager.UnregisterClient(userID)
			conn.Close()
			log.Printf("User %d disconnected", userID)
		}()

		// 循环读取消息
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			// 解析消息
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// 设置消息发送者和时间戳
			msg.FromUserID = userID
			msg.Timestamp = time.Now().Unix()

			// 处理不同类型的消息
			handleMessage(&msg)
		}
	})

	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
}

// handleMessage 处理不同类型的消息
func handleMessage(msg *Message) {
	switch msg.Type {
	case 1:
		// 私聊消息
		handlePrivateMessage(msg)
	case 2:
		// 获取私聊历史
		handleGetPrivateHistory(msg)
	case 3:
		// 群聊消息
		handleGroupMessage(msg)
	case 4:
		// 获取群聊历史
		handleGetGroupHistory(msg)
	default:
		log.Printf("Unknown message type: %d", msg.Type)
	}
}

// handlePrivateMessage 处理私聊消息
func handlePrivateMessage(msg *Message) {
	// 查找接收者的连接
	client, ok := manager.GetClient(msg.ToUserID)
	if !ok {
		log.Printf("User %d not connected", msg.ToUserID)
		return
	}

	// 发送消息给接收者
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// handleGetPrivateHistory 处理获取私聊历史请求
func handleGetPrivateHistory(msg *Message) {
	// 这里应该从数据库中获取历史消息
	// 为了演示，我们返回一个空的历史记录
	history := ChatHistory{
		Messages: []Message{},
		Total:    0,
	}

	// 设置响应数据
	msg.Data = history

	// 发送响应给请求者
	client, ok := manager.GetClient(msg.FromUserID)
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// handleGroupMessage 处理群聊消息
func handleGroupMessage(msg *Message) {
	// 获取聊天室中的所有用户
	userIDs := manager.GetRoomUsers(msg.RoomID)

	// 发送消息给聊天室中的所有用户
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, userID := range userIDs {
		client, ok := manager.GetClient(userID)
		if !ok {
			continue
		}

		if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

// handleGetGroupHistory 处理获取群聊历史请求
func handleGetGroupHistory(msg *Message) {
	// 这里应该从数据库中获取历史消息
	// 为了演示，我们返回一个空的历史记录
	history := ChatHistory{
		Messages: []Message{},
		Total:    0,
	}

	// 设置响应数据
	msg.Data = history

	// 发送响应给请求者
	client, ok := manager.GetClient(msg.FromUserID)
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
