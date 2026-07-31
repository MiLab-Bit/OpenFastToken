package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 允许向 WebSocket 写入消息的时间
	writeWait = 10 * time.Second

	// 允许从 WebSocket 读取下一条 pong 消息的时间
	pongWait = 60 * time.Second

	// 向客户端发送 ping 的间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 客户端发送消息的最大尺寸
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应限制 Origin
	},
}

// ReadPump 从 WebSocket 读取消息（客户端 -> 服务器）
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				SysError("WebSocket error: " + err.Error())
			}
			break
		}

		// 处理客户端发来的消息（如订阅通知类型等）
		handleClientMessage(c, message)
	}
}

// WritePump 向 WebSocket 写入消息（服务器 -> 客户端）
func (c *WebSocketClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了 channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中积压的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(bytes.NewBufferString("\n").Bytes())
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage 处理客户端发来的消息
func handleClientMessage(client *WebSocketClient, message []byte) {
	// 客户端可以发送 JSON 消息来订阅特定类型的通知
	// 例如: {"action": "subscribe", "types": ["billing", "balance"]}
	// 当前版本：记录收到的消息，后续可扩展订阅逻辑
	SysLog(fmt.Sprintf("WebSocket received from user %d: %s", client.userID, string(message)))
}

// ServeWs 处理 WebSocket 升级请求（给 HTTP 路由调用）
func ServeWs(hub *WebSocketHub, w http.ResponseWriter, r *http.Request, userID int, token string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	client := &WebSocketClient{
		hub:    hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		userID:  userID,
		token:   token,
	}
	client.hub.register <- client

	// 发送欢迎消息
	welcome := WebSocketNotification{
		Type:      "welcome",
		Title:     "Connected",
		Message:   "WebSocket connected successfully",
		Timestamp:  time.Now().Unix(),
		UserID:    userID,
		Level:     "info",
	}
	data, _ := json.Marshal(welcome)
	client.send <- data

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
