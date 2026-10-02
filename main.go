package main

import (
	"encoding/json"
	"log"
	"net/http"
	"omiro/middleware"
	"omiro/redis"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	serverID = uuid.NewString()
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	HandshakeTimeout:  10 * time.Second,
	Subprotocols:      []string{"chat"},
	EnableCompression: true,
}

func main() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	pass := os.Getenv("REDIS_PASS")
	if pass == "" {
		pass = ""
	}
	if err := redis.Init(redis.Config{
		Host:     host,
		Port:     port,
		Password: pass,
	}); err != nil {
		log.Fatal(err)
	}
	redis.RegisterServer(serverID)
	redis.StartSignalSubscriber(serverID, deliverToClient)
	e := echo.New()
	e.GET("/ws", func(c echo.Context) error {
		handleWebSocket(c.Response(), c.Request())
		return nil
	})

	e.GET("/session/new", func(c echo.Context) error {
		token, _, err := middleware.GenerateSessionToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session token"})
		}
		return c.JSON(http.StatusOK, map[string]any{"token": token})
	})
	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})
	go redis.StartMatchmaker()
	listen := os.Getenv("PORT")
	if listen == "" {
		listen = "8080"
	}
	e.Logger.Fatal(e.Start(":" + listen))
}

func deliverToClient(userID string, payload json.RawMessage) {
	clientsMu.RLock()
	client := clients[userID]
	clientsMu.RUnlock()

	if client == nil {
		return
	}

	client.trySend(websocket.TextMessage, payload)
}
