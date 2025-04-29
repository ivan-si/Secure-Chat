// api.js - Service for communicating with the Go backend

class APIService {
  // Get current identity of this chatter
  async getIdentity() {
    try {
      const response = await fetch('/api/identity');
      if (!response.ok) {
        throw new Error(`Failed to get identity: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error getting identity:', error);
      throw error;
    }
  }
  
  // Create a new identity
  async createIdentity() {
    try {
      const response = await fetch('/api/identity', {
        method: 'POST'
      });
      if (!response.ok) {
        throw new Error(`Failed to create identity: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error creating identity:', error);
      throw error;
    }
  }
  
  // Get list of chat partners
  async getPartners() {
    try {
      const response = await fetch('/api/partners');
      if (!response.ok) {
        throw new Error(`Failed to get partners: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error getting partners:', error);
      throw error;
    }
  }
  
  // Initiate handshake with a new partner
  async initiateHandshake(partnerKey) {
    try {
      const response = await fetch('/api/handshake/initiate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ partnerKey }),
      });
      if (!response.ok) {
        throw new Error(`Handshake failed: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error initiating handshake:', error);
      throw error;
    }
  }
  
  // Accept handshake from partner
  async returnHandshake(partnerKey, ephemeralKey) {
    try {
      const response = await fetch('/api/handshake/return', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          partnerKey, 
          ephemeralKey 
        }),
      });
      if (!response.ok) {
        throw new Error(`Return handshake failed: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error returning handshake:', error);
      throw error;
    }
  }
  
  // Finalize handshake as initiator
  async finalizeHandshake(partnerKey, ephemeralKey) {
    try {
      const response = await fetch('/api/handshake/finalize', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          partnerKey, 
          ephemeralKey 
        }),
      });
      if (!response.ok) {
        throw new Error(`Finalize handshake failed: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error finalizing handshake:', error);
      throw error;
    }
  }
  
  // Send a message to a partner
  async sendMessage(partnerKey, plaintext) {
    try {
      const response = await fetch('/api/message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          partnerKey,
          plaintext
        }),
      });
      if (!response.ok) {
        throw new Error(`Failed to send message: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error sending message:', error);
      throw error;
    }
  }
  
  // Receive a message
  async receiveMessage(message) {
    try {
      const response = await fetch('/api/message/receive', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message }),
      });
      if (!response.ok) {
        throw new Error(`Failed to receive message: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error receiving message:', error);
      throw error;
    }
  }
  
  // End a session with a partner
  async endSession(partnerKey) {
    try {
      const response = await fetch(`/api/session/${encodeURIComponent(partnerKey)}`, {
        method: 'DELETE'
      });
      if (!response.ok) {
        throw new Error(`Failed to end session: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error ending session:', error);
      throw error;
    }
  }
}

export default new APIService();
