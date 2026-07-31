package common

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketHub 管理所有 WebSocket 连接
// 用于向客户端推送实时通知（余额变化、请求日志等）
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan []byte
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

// WebSocketClient 单个 WebSocket 连接
type WebSocketClient struct {
	hub  *WebSocketHub
	conn *websocket.Conn
	send chan []byte
	userID int
	token string
}

var GlobalWebSocketHub *WebSocketHub

// InitWebSocketHub 初始化全局 WebSocket Hub
func InitWebSocketHub() {
	GlobalWebSocketHub = &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
	go GlobalWebSocketHub.Run()
}

// Run 启动 Hub 事件循环
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			SysLog(fmt.Sprintf("WebSocket client registered, total: %d", len(h.clients)))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			SysLog(fmt.Sprintf("WebSocket client unregistered, total: %d", len(h.clients)))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToAll 向所有连接的客户端广播消息
func (h *WebSocketHub) BroadcastToAll(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		SysError("WebSocket broadcast marshal error: " + err.Error())
		return
	}
	h.broadcast <- data
}

// BroadcastToUser 向指定用户推送消息
func (h *WebSocketHub) BroadcastToUser(userID int, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		SysError("WebSocket broadcast marshal error: " + err.Error())
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- data:
			default:
				// 客户端发送队列满，关闭连接
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// SendNotification 发送通知消息（给所有客户端或指定用户）
func (h *WebSocketHub) SendNotification(notification WebSocketNotification) {
	if notification.UserID > 0 {
		h.BroadcastToUser(notification.UserID, notification)
	} else {
		h.BroadcastToAll(notification)
	}
}

// WebSocketNotification 推送的通知结构
type WebSocketNotification struct {
	Type      string      `json:"type"`       // "billing", "log", "balance", "channel_status"
	Title     string      `json:"title"`      // 通知标题
	Message   string      `json:"message"`    // 通知内容
	Data      interface{} `json:"data"`       // 附加数据
	Timestamp  int64       `json:"timestamp"`   // 时间戳
	UserID    int         `json:"user_id"`    // 目标用户ID（0=所有用户）
	Level     string      `json:"level"`      // "info", "warning", "error"
}

// ClientCount 获取当前连接数
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
