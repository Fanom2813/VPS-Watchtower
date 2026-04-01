package protocol

import "encoding/json"

// Message types for WebSocket communication.
const (
	TypeAuthPair           = "auth:pair"
	TypeAuthPairSuccess    = "auth:pair:success"
	TypeAuthPairError      = "auth:pair:error"
	TypeAuthConnect        = "auth:connect"
	TypeAuthConnectSuccess = "auth:connect:success"
	TypeAuthConnectError   = "auth:connect:error"
)

// Message is the base envelope for all WebSocket messages.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// PairPayload is sent by the desktop to initiate pairing.
type PairPayload struct {
	PairingToken string `json:"pairingToken"`
}

// PairSuccessPayload is returned by the agent after successful pairing.
type PairSuccessPayload struct {
	Token string    `json:"token"`
	Agent AgentInfo `json:"agent"`
}

// ConnectPayload is sent by the desktop to authenticate with a stored token.
type ConnectPayload struct {
	Token string `json:"token"`
}

// ConnectSuccessPayload is returned by the agent after successful authentication.
type ConnectSuccessPayload struct {
	Agent AgentInfo `json:"agent"`
}

// AgentInfo contains the agent's identity and system information.
type AgentInfo struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Distro   string `json:"distro"`
	Version  string `json:"version"`
}

// ErrorPayload is returned on auth failure.
type ErrorPayload struct {
	Message string `json:"message"`
}

// Encode creates a Message with the given type and marshaled payload.
func Encode(msgType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{
		Type:    msgType,
		Payload: raw,
	})
}

// Decode unmarshals raw bytes into a Message envelope.
func Decode(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}

// DecodePayload unmarshals a Message's Payload into the target struct.
func DecodePayload(msg Message, target any) error {
	return json.Unmarshal(msg.Payload, target)
}
