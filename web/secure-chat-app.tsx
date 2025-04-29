import { useState, useEffect } from 'react';

const SecureChatApp = () => {
  const [identity, setIdentity] = useState(null);
  const [partners, setPartners] = useState([]);
  const [activePartner, setActivePartner] = useState(null);
  const [messages, setMessages] = useState({});
  const [newMessage, setNewMessage] = useState('');
  const [partnerKey, setPartnerKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Initialize chatter by fetching identity or creating new one
    const fetchIdentity = async () => {
      setLoading(true);
      try {
        const response = await fetch('/api/identity');
        if (response.ok) {
          const data = await response.json();
          setIdentity(data);
        } else {
          // Create new identity if none exists
          const createResponse = await fetch('/api/identity', {
            method: 'POST'
          });
          if (createResponse.ok) {
            const data = await createResponse.json();
            setIdentity(data);
          } else {
            throw new Error('Failed to create identity');
          }
        }
        
        // Fetch existing partners
        const partnersResponse = await fetch('/api/partners');
        if (partnersResponse.ok) {
          const partnersData = await partnersResponse.json();
          setPartners(partnersData);
        }
      } catch (err) {
        setError('Failed to initialize: ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchIdentity();
  }, []);

  const initiateHandshake = async () => {
    if (!partnerKey.trim()) {
      setError('Please enter a partner public key');
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch('/api/handshake/initiate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ partnerKey }),
      });
      
      if (!response.ok) {
        throw new Error('Handshake failed');
      }
      
      const data = await response.json();
      
      // Add partner to list if not already present
      if (!partners.find(p => p.key === partnerKey)) {
        const newPartner = {
          key: partnerKey,
          fingerprint: data.fingerprint,
          name: `Partner ${partners.length + 1}`,
        };
        setPartners([...partners, newPartner]);
        setActivePartner(newPartner);
        
        // Initialize empty message list for this partner
        setMessages(prev => ({
          ...prev,
          [partnerKey]: []
        }));
      }
      
      setPartnerKey('');
      setError(null);
    } catch (err) {
      setError('Failed to initiate handshake: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const sendMessage = async () => {
    if (!activePartner || !newMessage.trim()) {
      return;
    }
    
    setLoading(true);
    try {
      const response = await fetch('/api/message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          partnerKey: activePartner.key,
          plaintext: newMessage
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to send message');
      }
      
      // Add message to UI immediately
      const timestamp = new Date().toISOString();
      const newMsg = {
        text: newMessage,
        sender: 'me',
        timestamp
      };
      
      setMessages(prev => ({
        ...prev,
        [activePartner.key]: [...(prev[activePartner.key] || []), newMsg]
      }));
      
      setNewMessage('');
    } catch (err) {
      setError('Failed to send message: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const endSession = async () => {
    if (!activePartner) return;
    
    setLoading(true);
    try {
      const response = await fetch(`/api/session/${encodeURIComponent(activePartner.key)}`, {
        method: 'DELETE'
      });
      
      if (!response.ok) {
        throw new Error('Failed to end session');
      }
      
      // Remove partner from list
      setPartners(partners.filter(p => p.key !== activePartner.key));
      
      // Remove messages
      const newMessages = {...messages};
      delete newMessages[activePartner.key];
      setMessages(newMessages);
      
      setActivePartner(null);
    } catch (err) {
      setError('Failed to end session: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const selectPartner = (partner) => {
    setActivePartner(partner);
    setError(null);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(
      () => {
        // Flash success message
        const oldError = error;
        setError('Copied to clipboard!');
        setTimeout(() => setError(oldError), 2000);
      },
      () => {
        setError('Failed to copy to clipboard');
      }
    );
  };

  // Mock function to simulate receiving messages
  // In a real app, this would be replaced with a WebSocket or polling
  const mockReceiveMessage = () => {
    if (!activePartner) return;
    
    const mockMessage = {
      text: `This is a simulated response at ${new Date().toLocaleTimeString()}`,
      sender: 'them',
      timestamp: new Date().toISOString()
    };
    
    setMessages(prev => ({
      ...prev,
      [activePartner.key]: [...(prev[activePartner.key] || []), mockMessage]
    }));
  };

  if (loading && !identity) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-4 text-gray-700">Initializing secure chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-blue-600 text-white p-4 shadow-md">
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-xl font-semibold">Secure Chat</h1>
          {identity && (
            <div className="flex items-center">
              <span className="hidden md:inline mr-2">Your Identity:</span>
              <span className="text-sm bg-blue-700 rounded px-2 py-1 cursor-pointer" 
                    onClick={() => copyToClipboard(identity.publicKey)}>
                {identity.fingerprint}
              </span>
            </div>
          )}
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 bg-gray-200 flex flex-col border-r border-gray-300">
          <div className="p-4 border-b border-gray-300">
            <div className="flex flex-col space-y-2">
              <input
                type="text"
                placeholder="Partner's Public Key"
                className="p-2 border rounded text-sm"
                value={partnerKey}
                onChange={(e) => setPartnerKey(e.target.value)}
              />
              <button
                className="bg-blue-500 text-white p-2 rounded hover:bg-blue-600 text-sm"
                onClick={initiateHandshake}
                disabled={loading}
              >
                Start New Chat
              </button>
            </div>
          </div>
          
          <div className="flex-1 overflow-y-auto">
            <h2 className="px-4 py-2 text-sm font-semibold text-gray-600">Conversations</h2>
            {partners.length === 0 ? (
              <p className="px-4 py-2 text-sm text-gray-500">No conversations yet</p>
            ) : (
              <ul>
                {partners.map((partner) => (
                  <li 
                    key={partner.key}
                    className={`px-4 py-3 cursor-pointer hover:bg-gray-300 flex items-center ${
                      activePartner && activePartner.key === partner.key ? 'bg-gray-300' : ''
                    }`}
                    onClick={() => selectPartner(partner)}
                  >
                    <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-white mr-2">
                      {partner.name[0].toUpperCase()}
                    </div>
                    <div className="overflow-hidden">
                      <p className="font-medium truncate">{partner.name}</p>
                      <p className="text-xs text-gray-600 truncate">{partner.fingerprint}</p>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Chat Area */}
        <div className="flex-1 flex flex-col bg-white">
          {activePartner ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b flex justify-between items-center bg-gray-50">
                <div className="flex items-center">
                  <div className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center text-white mr-3">
                    {activePartner.name[0].toUpperCase()}
                  </div>
                  <div>
                    <h2 className="font-medium">{activePartner.name}</h2>
                    <p className="text-xs text-gray-600">{activePartner.fingerprint}</p>
                  </div>
                </div>
                <button 
                  className="px-3 py-1 rounded bg-red-100 text-red-700 hover:bg-red-200 text-sm"
                  onClick={endSession}
                >
                  End Session
                </button>
              </div>
              
              {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4" style={{ backgroundColor: '#f5f7fb' }}>
                {(!messages[activePartner.key] || messages[activePartner.key].length === 0) ? (
                  <div className="flex h-full items-center justify-center text-gray-500">
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {messages[activePartner.key].map((msg, i) => (
                      <div 
                        key={i} 
                        className={`flex ${msg.sender === 'me' ? 'justify-end' : 'justify-start'}`}
                      >
                        <div 
                          className={`max-w-xs md:max-w-md rounded-lg px-4 py-2 ${
                            msg.sender === 'me' 
                              ? 'bg-blue-600 text-white rounded-br-none' 
                              : 'bg-gray-300 text-gray-800 rounded-bl-none'
                          }`}
                        >
                          <p>{msg.text}</p>
                          <p className={`text-xs mt-1 ${msg.sender === 'me' ? 'text-blue-200' : 'text-gray-600'}`}>
                            {new Date(msg.timestamp).toLocaleTimeString()}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
              
              {/* Message Input */}
              <div className="p-4 border-t border-gray-300">
                <div className="flex">
                  <input
                    type="text"
                    placeholder="Type a message..."
                    className="flex-1 border rounded-l p-2"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                  />
                  <button 
                    className="bg-blue-600 text-white px-4 rounded-r hover:bg-blue-700"
                    onClick={sendMessage}
                  >
                    Send
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center bg-gray-50 text-gray-500">
              <div className="text-center max-w-md mx-auto p-6">
                <h2 className="text-xl font-semibold mb-2">No conversation selected</h2>
                <p className="mb-4">
                  Start a new chat by entering a partner's public key, or select an existing conversation.
                </p>
                {partners.length > 0 && (
                  <p className="text-sm border-t pt-4 text-gray-400">
                    You have {partners.length} existing conversation{partners.length !== 1 ? 's' : ''}.
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
      
      {/* Error Message */}
      {error && (
        <div className={`p-4 ${error === 'Copied to clipboard!' ? 'bg-green-500' : 'bg-red-500'} text-white fixed bottom-0 right-0 m-4 rounded shadow-lg`}>
          {error}
        </div>
      )}
      
      {/* Testing controls - these would be removed in a real app */}
      {activePartner && (
        <div className="fixed bottom-4 left-4 bg-yellow-100 p-2 rounded shadow-md border border-yellow-300">
          <button 
            className="text-xs px-2 py-1 bg-yellow-500 text-white rounded hover:bg-yellow-600"
            onClick={mockReceiveMessage}
          >
            Simulate Received Message
          </button>
        </div>
      )}
    </div>
  );
};

export default SecureChatApp;
