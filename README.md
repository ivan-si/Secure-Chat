# Secure Chat - End-to-End Encrypted Messaging

A forward-secure, end-to-end encrypted messaging platform that supports key compromise recovery and out-of-order message delivery. This project provides a robust implementation of secure communication primitives with a modern web interface.

## 🔒 Security Features

- **Forward Secrecy**: Messages remain secure even if keys are compromised later
- **End-to-End Encryption**: Using AES-GCM for authenticated encryption
- **Key Compromise Recovery**: Ratcheting mechanism to recover from compromised keys
- **Out-of-Order Message Handling**: Correctly processes messages received in any order
- **Diffie-Hellman Key Exchange**: Secure ephemeral key negotiation using NIST P-256 curve

## ⚠️ Security Notice

This code is primarily intended for educational purposes and may contain vulnerabilities or other bugs. Please do not use it for security-critical applications without a thorough security audit.

## 🏗️ Architecture

The system consists of several components:

1. **Cryptographic Core**: Implementations of ECDH and AES-GCM for secure key exchange and encryption
2. **Protocol Layer**: The messaging protocol implementation with forward secrecy and ratcheting
3. **Go API**: RESTful API for interacting with the backend cryptographic functions
4. **Web Interface**: React-based UI for end users to send and receive encrypted messages
5. **WebSocket Handler**: Real-time communication between clients

## 🚀 Getting Started

### Prerequisites

- Go 1.16+
- Node.js 14+
- npm or yarn

### Installation

```bash
# Clone the repository
git clone https://github.com/ivan-si/Secure-Chat.git
cd secure-chat

# Install Go dependencies
go mod tidy

# Install frontend dependencies
cd web
npm install
cd ..
```

### Running the Application

```bash
# Start the Go backend
go run cmd/secure-chat/main.go

# In another terminal, start the frontend development server
cd web
npm start
```

The application should now be available at http://localhost:3000

## 🧪 Testing

This project includes comprehensive test suites for all cryptographic components:

```bash
# Run all tests
go test ./...

# Run specific test suite
go test ./internal/crypto
go test ./internal/chat
```

## 📚 API Documentation

The Go API exposes the following endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/identity` | GET | Get user identity |
| `/identity` | POST | Create new identity |
| `/partners` | GET | List chat partners |
| `/handshake/initiate/:partner` | POST | Initiate handshake with partner |
| `/handshake/return/:partner` | POST | Return handshake from partner |
| `/handshake/finalize/:partner` | POST | Finalize handshake |
| `/message/send/:partner` | POST | Send encrypted message |
| `/message/receive` | POST | Receive and decrypt message |
| `/session/end/:partner` | POST | End session with partner |

## 🧩 Core Components

### Cryptographic Libraries

- **AESGCM.go**: Implementation of authenticated encryption/decryption with AES-GCM
- **ECDH.go**: Diffie-Hellman key generation and exchange using NIST P-256 curve

### Messaging Protocol

- **chatter.go**: Forward-secure messaging client with key compromise recovery

### API and Web Interface

- **go-api.go**: RESTful API for interacting with the core chat functionality
- **websocket-handler.go**: WebSocket implementation for real-time communication
- **secure-chat-app.tsx**: React-based user interface
- **api-service.js**: JavaScript service for connecting UI to the backend

## 🛠️ Development

### Project Structure

```
secure-chat/
├── cmd/
│   └── secure-chat/        # Application entry points
├── internal/
│   ├── crypto/             # Cryptographic primitives
│   ├── chat/               # Core messaging protocol
│   └── api/                # API handlers
└── web/                    # Frontend code
```

## 📝 License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## 🙏 Acknowledgments

- This project was inspired by the Signal Protocol
