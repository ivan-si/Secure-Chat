// Implementation of a forward-secure, end-to-end encrypted messaging client
// supporting key compromise recovery and out-of-order message delivery.
//
// SECURITY WARNING: This code is meant for educational purposes and may
// contain vulnerabilities or other bugs. Please do not use it for
// security-critical applications.

package main

import (
	//"bytes" //un-comment for helpers like bytes.equal
	"encoding/binary"
	"errors"
	//"fmt" //un-comment if you want to do any debug printing.
)

// Labels for key derivation

// Label for generating a check key from the initial root.
// Used for verifying the results of a handshake out-of-band.
const HANDSHAKE_CHECK_LABEL byte = 0x11

// Label for ratcheting the root key after deriving a key chain from it
const ROOT_LABEL = 0x22

// Label for ratcheting the main chain of keys
const CHAIN_LABEL = 0x33

// Label for deriving message keys from chain keys
const KEY_LABEL = 0x44

// Chatter represents a chat participant. Each Chatter has a single long-term
// key Identity, and a map of open sessions with other users (indexed by their
// identity keys). You should not need to modify this.
type Chatter struct {
	Identity *KeyPair
	Sessions map[PublicKey]*Session
}

// Session represents an open session between one chatter and another.
// You should not need to modify this, though you can add additional fields
// if you want to.
type Session struct {
	MyDHRatchet       *KeyPair
	PartnerDHRatchet  *PublicKey
	RootChain         *SymmetricKey
	SendChain         *SymmetricKey
	ReceiveChain      *SymmetricKey
	CachedReceiveKeys map[int]*SymmetricKey
	SendCounter       int
	LastUpdate        int
	ReceiveCounter    int
}

// Message represents a message as sent over an untrusted network.
// The first 5 fields are send unencrypted (but should be authenticated).
// The ciphertext contains the (encrypted) communication payload.
// You should not need to modify this.
type Message struct {
	Sender        *PublicKey
	Receiver      *PublicKey
	NextDHRatchet *PublicKey
	Counter       int
	LastUpdate    int
	Ciphertext    []byte
	IV            []byte
}

// EncodeAdditionalData encodes all of the non-ciphertext fields of a message
// into a single byte array, suitable for use as additional authenticated data
// in an AEAD scheme. You should not need to modify this code.
func (m *Message) EncodeAdditionalData() []byte {
	buf := make([]byte, 8+3*FINGERPRINT_LENGTH)

	binary.LittleEndian.PutUint32(buf, uint32(m.Counter))
	binary.LittleEndian.PutUint32(buf[4:], uint32(m.LastUpdate))

	if m.Sender != nil {
		copy(buf[8:], m.Sender.Fingerprint())
	}
	if m.Receiver != nil {
		copy(buf[8+FINGERPRINT_LENGTH:], m.Receiver.Fingerprint())
	}
	if m.NextDHRatchet != nil {
		copy(buf[8+2*FINGERPRINT_LENGTH:], m.NextDHRatchet.Fingerprint())
	}

	return buf
}

// NewChatter creates and initializes a new Chatter object. A long-term
// identity key is created and the map of sessions is initialized.
// You should not need to modify this code.
func NewChatter() *Chatter {
	c := new(Chatter)
	c.Identity = GenerateKeyPair()
	c.Sessions = make(map[PublicKey]*Session)
	return c
}

// EndSession erases all data for a session with the designated partner.
// All outstanding key material should be zeroized and the session erased.
func (c *Chatter) EndSession(partnerIdentity *PublicKey) error {

	if _, exists := c.Sessions[*partnerIdentity]; !exists {
		return errors.New("Don't have that session open to tear down")
	}

	delete(c.Sessions, *partnerIdentity)

	// TODO: your code here to zeroize remaining state

	return nil
}

// InitiateHandshake prepares the first message sent in a handshake, containing
// an ephemeral DH share. The partner which calls this method is the initiator.
func (c *Chatter) InitiateHandshake(partnerIdentity *PublicKey) (*PublicKey, error) {

	if _, exists := c.Sessions[*partnerIdentity]; exists {
		return nil, errors.New("Already have session open")
	}

	c.Sessions[*partnerIdentity] = &Session{
		CachedReceiveKeys: make(map[int]*SymmetricKey),
		// TODO: your code here
		MyDHRatchet: GenerateKeyPair(),
		SendCounter: 0,
	}

	// TODO: your code here

	c.Sessions[*partnerIdentity].SendChain = c.Sessions[*partnerIdentity].RootChain
	return &c.Sessions[*partnerIdentity].MyDHRatchet.PublicKey, nil
}

// ReturnHandshake prepares the second message sent in a handshake, containing
// an ephemeral DH share. The partner which calls this method is the responder.
func (c *Chatter) ReturnHandshake(partnerIdentity,
	partnerEphemeral *PublicKey) (*PublicKey, *SymmetricKey, error) {

	if _, exists := c.Sessions[*partnerIdentity]; exists {
		return nil, nil, errors.New("Already have session open")
	}

	c.Sessions[*partnerIdentity] = &Session{
		CachedReceiveKeys: make(map[int]*SymmetricKey),
		// TODO: your code here
		MyDHRatchet:      GenerateKeyPair(),
		PartnerDHRatchet: partnerEphemeral,
		SendCounter:      0,
	}

	// TODO: your code here
	rootChain := CombineKeys(
		DHCombine(partnerIdentity, &c.Sessions[*partnerIdentity].MyDHRatchet.PrivateKey),
		DHCombine(partnerEphemeral, &c.Identity.PrivateKey),
		DHCombine(partnerEphemeral, &c.Sessions[*partnerIdentity].MyDHRatchet.PrivateKey),
	)

	c.Sessions[*partnerIdentity].RootChain = rootChain
	//1st chain key derived from root key (Sender)
	//c.Sessions[*partnerIdentity].SendChain = c.Sessions[*partnerIdentity].RootChain.DeriveKey(CHAIN_LABEL)

	return &c.Sessions[*partnerIdentity].MyDHRatchet.PublicKey, rootChain.DeriveKey(HANDSHAKE_CHECK_LABEL), nil
}

// FinalizeHandshake lets the initiator receive the responder's ephemeral key
// and finalize the handshake.The partner which calls this method is the initiator.
func (c *Chatter) FinalizeHandshake(partnerIdentity,
	partnerEphemeral *PublicKey) (*SymmetricKey, error) {

	if _, exists := c.Sessions[*partnerIdentity]; !exists {
		return nil, errors.New("Can't finalize session, not yet open")
	}

	// TODO: your code here
	rootChain := CombineKeys(
		DHCombine(partnerEphemeral, &c.Identity.PrivateKey),
		DHCombine(partnerIdentity, &c.Sessions[*partnerIdentity].MyDHRatchet.PrivateKey),
		DHCombine(partnerEphemeral, &c.Sessions[*partnerIdentity].MyDHRatchet.PrivateKey),
	)
	c.Sessions[*partnerIdentity].RootChain = rootChain
	//1st chain key derived from root key (Reciever)
	//c.Sessions[*partnerIdentity].ReceiveChain = c.Sessions[*partnerIdentity].RootChain.DeriveKey(CHAIN_LABEL)

	return c.Sessions[*partnerIdentity].RootChain.DeriveKey(HANDSHAKE_CHECK_LABEL), nil
}

// SendMessage is used to send the given plaintext string as a message.
// You'll need to implement the code to ratchet, derive keys and encrypt this message.
func (c *Chatter) SendMessage(partnerIdentity *PublicKey,
	plaintext string) (*Message, error) {

	if _, exists := c.Sessions[*partnerIdentity]; !exists {
		return nil, errors.New("Can't send message to partner with no open session")
	}

	message := &Message{
		Sender:   &c.Identity.PublicKey,
		Receiver: partnerIdentity,
		// TODO: your code here
		IV:      NewIV(), // Generate a new IV
		Counter: 0,
	}
	chainKey := c.Sessions[*partnerIdentity].RootChain.DeriveKey(CHAIN_LABEL)
	c.Sessions[*partnerIdentity].RootChain.Zeroize()
	c.Sessions[*partnerIdentity].RootChain = chainKey
	// TODO: your code here
	//Increase send counter
	c.Sessions[*partnerIdentity].SendCounter += 1
	// Current message key
	messageKey := c.Sessions[*partnerIdentity].SendChain.DeriveKey(KEY_LABEL)
	// encrypts plaintext and assignes it to message.ciphertext
	message.Ciphertext = messageKey.AuthenticatedEncrypt(plaintext, nil, message.IV)
	return message, nil
}

// ReceiveMessage is used to receive the given message and return the correct
// plaintext. This method is where most of the key derivation, ratcheting
// and out-of-order message handling logic happens.
func (c *Chatter) ReceiveMessage(message *Message) (string, error) {

	if _, exists := c.Sessions[*message.Sender]; !exists {
		return "", errors.New("Can't receive message from partner with no open session")
	}
	// TODO: your code here
	//message.Counter += 1
	// finds current message key and assigns value to messageKey
	messageKey := c.Sessions[*message.Sender].ReceiveChain.DeriveKey(KEY_LABEL)
	//assign decrypted text to plaintext var
	plaintext, _ := messageKey.AuthenticatedDecrypt(message.Ciphertext, nil, message.IV)

	return plaintext, nil
}
