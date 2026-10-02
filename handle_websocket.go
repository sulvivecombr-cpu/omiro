package main

import (
	"encoding/json"
	"log"
	"net/http"
	"omiro/middleware"
	"omiro/redis"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var clients = make(map[string]*Client)
var clientsMu sync.RWMutex

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !middleware.EnsureUpgradeChecks(w, r) {
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	defer conn.Close()
	log.Println("client connected:", conn.RemoteAddr())

	client := &Client{
		ID:   uuid.NewString(),
		Conn: conn,
		Send: make(chan SendMessageType, 256),
	}

	clientsMu.Lock()
	clients[client.ID] = client
	clientsMu.Unlock()

	ip := r.RemoteAddr
	if err := redis.RegisterClient(client.ID, ip, serverID); err != nil {
		log.Println("failed to register client in redis:", err)
		return
	}

	go client.writePump()
	go client.readPump()

	welcomeMsg := map[string]any{
		"message":   "Hello from server",
		"client_id": client.ID,
		"timestamp": time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(welcomeMsg)
	if err != nil {
		log.Println("json marshal error:", err)
		return
	}

	client.Send <- SendMessageType{
		Message: msgBytes,
		Type:    websocket.TextMessage,
	}

	select {}
}
