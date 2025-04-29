package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

func getIdentityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Return the public key of our Chatter
	response := IdentityResponse{
		PublicKey:   base64.StdEncoding.EncodeToString(myChatter.Identity.PublicKey.Serialize()),
		Fingerprint: base64.StdEncoding.EncodeToString(myChatter.Identity.PublicKey.Fingerprint()),
	}

	json.NewEncoder(w).Encode(response)
}

func createIdentityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// In a real application, you might want to save the previous identity
	// and ask for confirmation before creating a new one
	myChatter = NewChatter()

	// Clear all sessions and partners since we have a new identity
	mutex.Lock()
	chatSessions = make(map[string]*SessionInfo)
	partners = make(map[string]*PartnerInfo)
	mutex.Unlock()

	response := IdentityResponse{
		PublicKey:   base64.StdEncoding.EncodeToString(myChatter.Identity.PublicKey.Serialize()),
		Fingerprint: base64.StdEncoding.EncodeToString(myChatter.Identity.PublicKey.Fingerprint()),
	}

	json.NewEncoder(w).Encode(response)
}

func getPartnersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mutex.RLock()
	defer mutex.RUnlock()

	// Convert our partners map to a slice for the response
	partnersList := make([]map[string]string, 0, len(partners))

	for _, partner := range partners {
		partnersList = append(partnersList, map[string]string{
			"key":         base64.StdEncoding.EncodeToString(partner.PublicKey.Serialize()),
			"fingerprint": base64.StdEncoding.EncodeToString(partner.PublicKey.Fingerprint()),
			"name":        partner.Name,
		})
	}

	json.NewEncoder(w).Encode(partnersList)
}

func initiateHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req HandshakeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode the partner's public key
	partnerKeyBytes, err := base64.StdEncoding.DecodeString(req.PartnerKey)
	if err != nil {
		http.Error(w, "Invalid partner key format", http.StatusBadRequest)
		return
	}

	partnerKey := &PublicKey{}
	err = partnerKey.Deserialize(partnerKeyBytes)
	if err != nil {
		http.Error(w, "Invalid partner key", http.StatusBadRequest)
		return
	}

	// Initiate handshake
	ephemeralKey, err := myChatter.InitiateHandshake(partnerKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Handshake failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Store partner information
	partnerFingerprint := base64.StdEncoding.EncodeToString(partnerKey.Fingerprint())
	partnerName := fmt.Sprintf("Partner-%s", partnerFingerprint[:8])

	mutex.Lock()
	// Add to partners map if not already present
	if _, exists := partners[req.PartnerKey]; !exists {
		partners[req.PartnerKey] = &PartnerInfo{
			Name:        partnerName,
			PublicKey:   partnerKey,
			Fingerprint: partnerFingerprint,
		}

		// Initialize a session
		chatSessions[req.PartnerKey] = &SessionInfo{
			Partner:     partnerKey,
			Messages:    []ChatMessage{},
			Initialized: false, // Will be set to true after handshake completes
		}

		// Notify clients about the new session
		NotifySessionChange(req.PartnerKey, "created")
	}
	mutex.Unlock()

	// Notify clients about the handshake initiation
	NotifyHandshakeStatus(req.PartnerKey, "initiated")

	// Return the ephemeral key
	response := HandshakeResponse{
		EphemeralKey: base64.StdEncoding.EncodeToString(ephemeralKey.Serialize()),
		Fingerprint:  partnerFingerprint,
		Success:      true,
	}

	json.NewEncoder(w).Encode(response)
}

func returnHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req HandshakeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode keys
	partnerKeyBytes, err := base64.StdEncoding.DecodeString(req.PartnerKey)
	if err != nil {
		http.Error(w, "Invalid partner key format", http.StatusBadRequest)
		return
	}

	ephKeyBytes, err := base64.StdEncoding.DecodeString(req.EphemeralKey)
	if err != nil {
		http.Error(w, "Invalid ephemeral key format", http.StatusBadRequest)
		return
	}

	partnerKey := &PublicKey{}
	err = partnerKey.Deserialize(partnerKeyBytes)
	if err != nil {
		http.Error(w, "Invalid partner key", http.StatusBadRequest)
		return
	}

	ephKey := &PublicKey{}
	err = ephKey.Deserialize(ephKeyBytes)
	if err != nil {
		http.Error(w, "Invalid ephemeral key", http.StatusBadRequest)
		return
	}

	// Return handshake
	responseKey, err := myChatter.ReturnHandshake(partnerKey, ephKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Return handshake failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Store partner information
	partnerFingerprint := base64.StdEncoding.EncodeToString(partnerKey.Fingerprint())
	partnerName := fmt.Sprintf("Partner-%s", partnerFingerprint[:8])

	partnerKeyStr := req.PartnerKey

	mutex.Lock()
	// Add to partners map if not already present
	if _, exists := partners[partnerKeyStr]; !exists {
		partners[partnerKeyStr] = &PartnerInfo{
			Name:        partnerName,
			PublicKey:   partnerKey,
			Fingerprint: partnerFingerprint,
		}

		// Initialize a session
		chatSessions[partnerKeyStr] = &SessionInfo{
			Partner:     partnerKey,
			Messages:    []ChatMessage{},
			Initialized: true, // Responder completes handshake immediately
		}
	} else {
		// Update existing session
		chatSessions[partnerKeyStr].Initialized = true
	}
	mutex.Unlock()

	// Return the ephemeral key
	response := HandshakeResponse{
		EphemeralKey: base64.StdEncoding.EncodeToString(responseKey.Serialize()),
		Fingerprint:  partnerFingerprint,
		Success:      true,
	}

	json.NewEncoder(w).Encode(response)
}

func finalizeHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req HandshakeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode keys
	partnerKeyBytes, err := base64.StdEncoding.DecodeString(req.PartnerKey)
	if err != nil {
		http.Error(w, "Invalid partner key format", http.StatusBadRequest)
		return
	}

	ephKeyBytes, err := base64.StdEncoding.DecodeString(req.EphemeralKey)
	if err != nil {
		http.Error(w, "Invalid ephemeral key format", http.StatusBadRequest)
		return
	}

	partnerKey := &PublicKey{}
	err = partnerKey.Deserialize(partnerKeyBytes)
	if err != nil {
		http.Error(w, "Invalid partner key", http.StatusBadRequest)
		return
	}

	ephKey := &PublicKey{}
	err = ephKey.Deserialize(ephKeyBytes)
	if err != nil {
		http.Error(w, "Invalid ephemeral key", http.StatusBadRequest)
		return
	}

	// Finalize handshake
	err = myChatter.FinalizeHandshake(partnerKey, ephKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Finalize handshake failed: %v", err), http.StatusInternalServerError)
		return
	}

	partnerKeyStr := req.PartnerKey

	mutex.Lock()
	if session, exists := chatSessions[partnerKeyStr]; exists {
		session.Initialized = true
	}
	mutex.Unlock()

	// Notify clients that handshake is finalized
	NotifyHandshakeStatus(partnerKeyStr, "finalized")

	response := HandshakeResponse{
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req MessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode the partner's public key
	partnerKeyBytes, err := base64.StdEncoding.DecodeString(req.PartnerKey)
	if err != nil {
		http.Error(w, "Invalid partner key format", http.StatusBadRequest)
		return
	}

	partnerKey := &PublicKey{}
	err = partnerKey.Deserialize(partnerKeyBytes)
	if err != nil {
		http.Error(w, "Invalid partner key", http.StatusBadRequest)
		return
	}

	// Check if session is initialized
	mutex.RLock()
	session, exists := chatSessions[req.PartnerKey]
	if !exists || !session.Initialized {
		mutex.RUnlock()
		http.Error(w, "Session not initialized", http.StatusBadRequest)
		return
	}
	mutex.RUnlock()

	// Send the message
	message, err := myChatter.SendMessage(partnerKey, req.Plaintext)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Store message in session
	mutex.Lock()
	session.Messages = append(session.Messages, ChatMessage{
		Plaintext: req.Plaintext,
		Sender:    "me",
		Timestamp: message.Timestamp.String(),
	})
	mutex.Unlock()

	response := MessageResponse{
		Message: message,
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

func receiveMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req ReceiveMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == nil {
		http.Error(w, "Missing message", http.StatusBadRequest)
		return
	}

	// Process the message
	plaintext, err := myChatter.ReceiveMessage(req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to receive message: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the partner associated with this message
	senderKey := req.Message.Sender
	senderKeyStr := base64.StdEncoding.EncodeToString(senderKey.Serialize())

	// Store message in session
	mutex.Lock()
	if session, exists := chatSessions[senderKeyStr]; exists {
		chatMessage := ChatMessage{
			Plaintext: plaintext,
			Sender:    "them",
			Timestamp: req.Message.Timestamp.String(),
		}
		session.Messages = append(session.Messages, chatMessage)
	}
	mutex.Unlock()

	// Notify all clients about new message
	NotifyMessageReceived(senderKeyStr, req.Message, plaintext)

	response := MessageResponse{
		Message:   req.Message,
		Success:   true,
		Plaintext: plaintext,
	}

	json.NewEncoder(w).Encode(response)
}

func endSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	partnerKeyStr := vars["partnerKey"]

	// Decode the partner's public key
	partnerKeyBytes, err := base64.StdEncoding.DecodeString(partnerKeyStr)
	if err != nil {
		http.Error(w, "Invalid partner key format", http.StatusBadRequest)
		return
	}

	partnerKey := &PublicKey{}
	err = partnerKey.Deserialize(partnerKeyBytes)
	if err != nil {
		http.Error(w, "Invalid partner key", http.StatusBadRequest)
		return
	}

	// End the session
	err = myChatter.EndSession(partnerKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to end session: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove partner and session from our maps
	mutex.Lock()
	delete(partners, partnerKeyStr)
	delete(chatSessions, partnerKeyStr)
	mutex.Unlock()

	response := map[string]bool{"success": true}
	json.NewEncoder(w).Encode(response)
}
