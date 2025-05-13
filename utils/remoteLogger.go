// extension of the logger package to send logs to a remote server

package utils

import (
	"BlockChainSimulator/config"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type LogMessage struct {
	ShardID   int    `json:"shardId"`
	NodeID    int    `json:"nodeId"`
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}

func (l *Logger) initWebSocket() {
	if !config.ConnectRemoteDemo {
		fmt.Printf("No WebSocket\n")
		return
	}
	if l.wsConn != nil {
		// 检查连接是否仍然活跃
		if err := l.wsConn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err == nil {
			log.Println("WebSocket connection already exists and is active")
			return
		}
		// 连接已失效，关闭旧连接
		l.wsConn.Close()
		l.wsConn = nil
	}
	url := "ws://" + config.DemoServerURL + "/api/log"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v", err)
		return
	} else {
		fmt.Printf("Connected to WebSocket\n")
	}
	l.wsConn = conn
}

// 发送日志到UI服务器
func (l *Logger) sendLogToUI(level string, format string, v ...interface{}) {
	if !config.ConnectRemoteDemo {
		return
	}

	message := fmt.Sprintf(format, v...)
	source := l.getCallerInfo(4)
	timestamp := time.Now().Format("15:04:05.000")

	logMessage := LogMessage{
		ShardID:   l.shardID,
		NodeID:    l.nodeID,
		Level:     level,
		Timestamp: timestamp,
		Message:   message,
		Source:    source,
	}

	jsonData, err := json.Marshal(logMessage)
	if err != nil {
		log.Printf("Error marshalling log message: %v", err)
		return
	}

	err = l.wsConn.WriteMessage(websocket.TextMessage, jsonData)

	if err != nil {
		log.Printf("Error sending log message: %v", err)
	} else {
		fmt.Printf("Send log message: %v\n", jsonData)
	}
}
