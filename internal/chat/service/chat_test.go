package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gobili/config"
	"gobili/pkg/constants"
	"gobili/pkg/db"
	dbmodel "gobili/pkg/db/model"
	"gobili/pkg/token"
	"gobili/pkg/utils"
)

const (
	testDSNEnv   = "GOBILI_TEST_DSN"
	testConfig   = "../../../config/config.example.yaml"
	readDeadline = 5 * time.Second
)

// setupService 连测试库并构造聊天服务。未配置 DSN 时跳过。
func setupService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("未设置 %s，跳过聊天服务集成测试", testDSNEnv)
	}
	if err := config.Init(testConfig); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	sf, err := utils.NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("创建雪花生成器失败: %v", err)
	}

	return newService(db.NewDatabase(gormDB, sf), sf, nil), gormDB
}

// seedUsers 造两个用户，返回它们的 ID。
func seedUsers(t *testing.T, gormDB *gorm.DB) (int64, int64) {
	t.Helper()

	if err := gormDB.Exec("TRUNCATE TABLE " + constants.ChatMessageTableName).Error; err != nil {
		t.Fatalf("清空消息表失败: %v", err)
	}
	if err := gormDB.Exec("TRUNCATE TABLE " + constants.UserTableName).Error; err != nil {
		t.Fatalf("清空用户表失败: %v", err)
	}

	users := []*dbmodel.User{
		{ID: 1001, Username: "chat-a", PasswordHash: "h"},
		{ID: 1002, Username: "chat-b", PasswordHash: "h"},
	}
	if err := gormDB.Table(constants.UserTableName).Create(&users).Error; err != nil {
		t.Fatalf("造用户失败: %v", err)
	}
	return 1001, 1002
}

// dial 用访问令牌建立一条 WebSocket 连接，并读掉 connected 帧。
func dial(t *testing.T, serverURL string, userID int64) *websocket.Conn {
	t.Helper()

	access, _, err := token.CreateAllToken(userID)
	if err != nil {
		t.Fatalf("签发令牌失败: %v", err)
	}

	url := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws?token=" + access
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("建立 WebSocket 失败: %v", err)
	}
	_ = resp.Body.Close()

	if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
		t.Fatalf("设置读超时失败: %v", err)
	}
	var first outgoingMessage
	if err := conn.ReadJSON(&first); err != nil {
		t.Fatalf("读取握手帧失败: %v", err)
	}
	if first.Type != "connected" {
		t.Fatalf("首个帧应为 connected，实际为 %q", first.Type)
	}
	return conn
}

func readMessage(t *testing.T, conn *websocket.Conn) outgoingMessage {
	t.Helper()
	if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
		t.Fatalf("设置读超时失败: %v", err)
	}
	var msg outgoingMessage
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("读取消息失败: %v", err)
	}
	return msg
}

func TestChatDeliversAndPersistsMessage(t *testing.T) {
	svc, gormDB := setupService(t)
	sender, receiver := seedUsers(t, gormDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/history", svc.HandleHistory)
	mux.HandleFunc("/conversations", svc.HandleConversations)
	server := httptest.NewServer(mux)
	defer server.Close()

	connA := dial(t, server.URL, sender)
	defer func() { _ = connA.Close() }()
	connB := dial(t, server.URL, receiver)
	defer func() { _ = connB.Close() }()

	if err := connA.WriteJSON(incomingMessage{
		ToUserID: "1002", Content: "在吗",
	}); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	// 接收方应当实时收到
	got := readMessage(t, connB)
	if got.Type != "message" || got.Content != "在吗" {
		t.Fatalf("接收方收到的帧不对: %+v", got)
	}
	if got.FromUser != "1001" || got.ToUser != "1002" {
		t.Fatalf("收发双方不对: %+v", got)
	}
	if got.RoomID != "1001_1002" {
		t.Fatalf("会话 ID 应按双方 ID 排序拼接，实际 %q", got.RoomID)
	}

	// 发送方应当收到回执，且标记对方在线
	ack := readMessage(t, connA)
	if ack.Type != "ack" || !ack.Online {
		t.Fatalf("回执不对: %+v", ack)
	}

	// 消息必须落库，历史接口能拉到
	history := fetchHistory(t, server.URL, receiver, sender)
	if history.Data.Total != 1 || len(history.Data.Items) != 1 {
		t.Fatalf("历史消息数不对: %+v", history)
	}
	if history.Data.Items[0].Content != "在吗" {
		t.Fatalf("历史消息内容不对: %+v", history.Data.Items[0])
	}
}

// 收件人不存在时必须直接拒绝。
//
// 旧版完全没校验，任意 ID 都能发，会话表里会堆满指向不存在用户的消息。
func TestChatRejectsUnknownReceiver(t *testing.T) {
	svc, gormDB := setupService(t)
	sender, _ := seedUsers(t, gormDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	conn := dial(t, server.URL, sender)
	defer func() { _ = conn.Close() }()

	if err := conn.WriteJSON(incomingMessage{ToUserID: "999999", Content: "hi"}); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	got := readMessage(t, conn)
	if got.Type != "error" || !strings.Contains(got.Error, "接收方不存在") {
		t.Fatalf("应返回「接收方不存在」，实际 %+v", got)
	}

	var count int64
	if err := gormDB.Table(constants.ChatMessageTableName).Count(&count).Error; err != nil {
		t.Fatalf("统计消息失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("被拒绝的消息不该落库，实际写了 %d 条", count)
	}
}

func TestChatRejectsSelfMessage(t *testing.T) {
	svc, gormDB := setupService(t)
	sender, _ := seedUsers(t, gormDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	conn := dial(t, server.URL, sender)
	defer func() { _ = conn.Close() }()

	if err := conn.WriteJSON(incomingMessage{ToUserID: "1001", Content: "hi"}); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	if got := readMessage(t, conn); got.Type != "error" {
		t.Fatalf("给自己发消息应当被拒，实际 %+v", got)
	}
}

func TestChatRejectsBadOrigin(t *testing.T) {
	svc, gormDB := setupService(t)
	sender, _ := seedUsers(t, gormDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	access, _, err := token.CreateAllToken(sender)
	if err != nil {
		t.Fatalf("签发令牌失败: %v", err)
	}

	// 白名单为空时，带 Origin 的浏览器请求必须被拒
	header := http.Header{}
	header.Set("Origin", "http://evil.example.com")
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=" + access

	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err == nil {
		_ = conn.Close()
		t.Fatal("不在白名单的 Origin 应当握手失败")
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
}

func TestChatRejectsMissingToken(t *testing.T) {
	svc, _ := setupService(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		_ = conn.Close()
		t.Fatal("不带令牌应当握手失败")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("应当返回 401，实际 %+v", resp)
	}
	_ = resp.Body.Close()
}

// historyResponse 对应 /history 的响应结构。
type historyResponse struct {
	Code string `json:"code"`
	Data struct {
		Items []historyItem `json:"items"`
		Total int64         `json:"total"`
	} `json:"data"`
}

// fetchHistory 以 viewer 的身份拉取与 peer 的会话历史。
func fetchHistory(t *testing.T, serverURL string, viewer, peer int64) historyResponse {
	t.Helper()

	access, _, err := token.CreateAllToken(viewer)
	if err != nil {
		t.Fatalf("签发令牌失败: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		serverURL+"/history?peer_id="+strconv.FormatInt(peer, 10), nil)
	if err != nil {
		t.Fatalf("构造请求失败: %v", err)
	}
	req.Header.Set(constants.AuthHeader, access)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求历史失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body historyResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("解析历史响应失败: %v", err)
	}
	return body
}


func TestChatListsConversations(t *testing.T) {
	svc, gormDB := setupService(t)
	sender, receiver := seedUsers(t, gormDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/conversations", svc.HandleConversations)
	server := httptest.NewServer(mux)
	defer server.Close()

	connA := dial(t, server.URL, sender)
	defer func() { _ = connA.Close() }()
	connB := dial(t, server.URL, receiver)
	defer func() { _ = connB.Close() }()

	if err := connA.WriteJSON(incomingMessage{ToUserID: "1002", Content: "第一句"}); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	_ = readMessage(t, connB)
	_ = readMessage(t, connA)

	if err := connB.WriteJSON(incomingMessage{ToUserID: "1001", Content: "最新一句"}); err != nil {
		t.Fatalf("回复消息失败: %v", err)
	}
	_ = readMessage(t, connA)
	_ = readMessage(t, connB)

	got := fetchConversations(t, server.URL, receiver)
	if got.Data.Total != 1 || len(got.Data.Items) != 1 {
		t.Fatalf("会话列表数量不对: %+v", got)
	}
	item := got.Data.Items[0]
	if item.PeerID != "1001" {
		t.Fatalf("peer_id 应为对方 1001，实际 %q", item.PeerID)
	}
	if item.LastContent != "最新一句" {
		t.Fatalf("应展示最新消息，实际 %+v", item)
	}
}

type conversationsResponse struct {
	Code string `json:"code"`
	Data struct {
		Items []conversationItem `json:"items"`
		Total int64             `json:"total"`
	} `json:"data"`
}

func fetchConversations(t *testing.T, serverURL string, viewer int64) conversationsResponse {
	t.Helper()

	access, _, err := token.CreateAllToken(viewer)
	if err != nil {
		t.Fatalf("签发令牌失败: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		serverURL+"/conversations", nil)
	if err != nil {
		t.Fatalf("构造请求失败: %v", err)
	}
	req.Header.Set(constants.AuthHeader, access)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求会话列表失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body conversationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("解析会话列表失败: %v", err)
	}
	return body
}
