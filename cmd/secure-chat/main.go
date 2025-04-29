package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Global state for our application
var (
	myChatter    *Chatter
	chatSessions map[string]*SessionInfo
	partners     map[string]*PartnerInfo
	mutex        sync.RWMutex
)

// PartnerInfo contains information about a chat partner
type PartnerInfo struct {
	Name        string     `json:"name"`
	PublicKey   *PublicKey `json:"publicKey"`
	Fingerprint string     `json:"fingerprint"`
}

// SessionInfo tracks the state of a chat session
type SessionInfo struct {
	Partner     *PublicKey    `json:"partner"`
	Messages    []ChatMessage `json:"messages"`
	Initialized bool          `json:"initialized"`
}

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Plaintext string `json:"plaintext"`
	Sender    string `json:"sender"` // "me" or "them"
	Timestamp string `json:"timestamp"`
}

// API Request/Response structs
type IdentityResponse struct {
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
}

type HandshakeRequest struct {
	PartnerKey   string `json:"partnerKey"`
	EphemeralKey string `json:"ephemeralKey,omitempty"`
}

type HandshakeResponse struct {
	EphemeralKey string `json:"ephemeralKey"`
	Fingerprint  string `json:"fingerprint"`
	Success      bool   `json:"success"`
}

type MessageRequest struct {
	PartnerKey string `json:"partnerKey"`
	Plaintext  string `json:"plaintext"`
}

type MessageResponse struct {
	Message   *Message `json:"message"`
	Success   bool     `json:"success"`
	Plaintext string   `json:"plaintext,omitempty"`
}

type ReceiveMessageRequest struct {
	Message *Message `json:"message"`
}

func init() {
	// Initialize our application state
	myChatter = NewChatter()
	chatSessions = make(map[string]*SessionInfo)
	partners = make(map[string]*PartnerInfo)
}

func main() {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/identity", getIdentityHandler).Methods("GET")
	r.HandleFunc("/api/identity", createIdentityHandler).Methods("POST")
	r.HandleFunc("/api/partners", getPartnersHandler).Methods("GET")
	r.HandleFunc("/api/handshake/initiate", initiateHandshakeHandler).Methods("POST")
	r.HandleFunc("/api/handshake/return", returnHandshakeHandler).Methods("POST")
	r.HandleFunc("/api/handshake/finalize", finalizeHandshakeHandler).Methods("POST")
	r.HandleFunc("/api/message", sendMessageHandler).Methods("POST")
	r.HandleFunc("/api/message/receive", receiveMessageHandler).Methods("POST")
	r.HandleFunc("/api/session/{partnerKey}", endSessionHandler).Methods("DELETE")

	// WebSocket endpoint
	r.HandleFunc("/ws", HandleWebSocket)

	// Serve static files for the React app
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Start the server
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
