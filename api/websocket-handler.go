package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// Upgrader for WebSocket connections
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow all origins for development, in production you would restrict this
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// Map of client connections
	clients = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

// WebSocketEvent represents different types of events sent over the WebSocket
type WebSocketEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// MessageEvent is sent when a new message is received
type MessageEvent struct {
	SenderKey  string       `json:"senderKey"`
	Message    *Message     `json:"message"`
	Plaintext  string       `json:"plaintext"`
	ChatMessage ChatMessage `json:"chatMessage"`
}

// HandshakeEvent is sent when a handshake is initiated
type HandshakeEvent struct {
	PartnerKey string `json:"partnerKey"`
	Status     string `json:"status"` // "initiated", "returned", "finalized"
}

// SessionEvent is sent when a session state changes
type SessionEvent struct {
	PartnerKey string `json:"partnerKey"`
	Action     string `json:"action"` // "created", "ended"
}

// HandleWebSocket upgrades the HTTP connection to a WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()
	
	// Unregister client when function returns
	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
	}()

	// Read messages in a loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// We're only using WebSockets for server-to-client notifications 
		// so we don't need to process incoming messages here
	}
}

// BroadcastEvent sends an event to all connected clients
func BroadcastEvent(eventType string, payload interface{}) {
	event := WebSocketEvent{
		Type:    eventType,
		Payload: payload,
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}
	
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Failed to send message to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// NotifyMessageReceived broadcasts a message received event
func NotifyMessageReceived(senderKeyStr string, message *Message, plaintext string) {
	// Create a ChatMessage from the received message
	chatMessage := ChatMessage{
		Plaintext: plaintext,
		Sender:    "them",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	event := MessageEvent{
		SenderKey:   senderKeyStr,
		Message:     message,
		Plaintext:   plaintext,
		ChatMessage: chatMessage,
	}
	
	BroadcastEvent("message", event)
}

// NotifyHandshakeStatus broadcasts a handshake status change
func NotifyHandshakeStatus(partnerKey string, status string) {
	event := HandshakeEvent{
		PartnerKey: partnerKey,
		Status:     status,
	}
	
	BroadcastEvent("handshake", event)
}

// NotifySessionChange broadcasts a session state change
func NotifySessionChange(partnerKey string, action string) {
	event := SessionEvent{
		PartnerKey: partnerKey,
		Action:     action,
	}
	
	BroadcastEvent("session", event)
}
