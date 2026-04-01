package protocol_test

import (
	"encoding/json"
	"testing"

	"github.com/eyes-on-vps/agent/internal/protocol"
)

func TestEncodeDecode_PairMessage(t *testing.T) {
	payload := protocol.PairPayload{
		PairingToken: "tok-abc123",
		AgentID:      "agent-1",
		Hostname:     "vps-prod-01",
	}

	data, err := protocol.Encode(protocol.TypeAuthPair, payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthPair {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthPair, msg.Type)
	}

	var decoded protocol.PairPayload
	if err := protocol.DecodePayload(msg, &decoded); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if decoded.PairingToken != payload.PairingToken {
		t.Errorf("pairingToken: expected %q, got %q", payload.PairingToken, decoded.PairingToken)
	}
	if decoded.AgentID != payload.AgentID {
		t.Errorf("agentId: expected %q, got %q", payload.AgentID, decoded.AgentID)
	}
	if decoded.Hostname != payload.Hostname {
		t.Errorf("hostname: expected %q, got %q", payload.Hostname, decoded.Hostname)
	}
}

func TestEncodeDecode_ConnectMessage(t *testing.T) {
	payload := protocol.ConnectPayload{
		AgentToken: "jwt-token-here",
	}

	data, err := protocol.Encode(protocol.TypeAuthConnect, payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthConnect {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthConnect, msg.Type)
	}

	var decoded protocol.ConnectPayload
	if err := protocol.DecodePayload(msg, &decoded); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if decoded.AgentToken != payload.AgentToken {
		t.Errorf("agentToken: expected %q, got %q", payload.AgentToken, decoded.AgentToken)
	}
}

func TestEncodeDecode_PairSuccessPayload(t *testing.T) {
	payload := protocol.PairSuccessPayload{
		AgentToken: "new-jwt-token",
	}

	data, err := protocol.Encode(protocol.TypeAuthPairSuccess, payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthPairSuccess {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthPairSuccess, msg.Type)
	}

	var decoded protocol.PairSuccessPayload
	if err := protocol.DecodePayload(msg, &decoded); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if decoded.AgentToken != payload.AgentToken {
		t.Errorf("agentToken: expected %q, got %q", payload.AgentToken, decoded.AgentToken)
	}
}

func TestEncodeDecode_ErrorPayload(t *testing.T) {
	payload := protocol.ErrorPayload{
		Message: "invalid pairing token",
	}

	data, err := protocol.Encode(protocol.TypeAuthPairError, payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthPairError {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthPairError, msg.Type)
	}

	var decoded protocol.ErrorPayload
	if err := protocol.DecodePayload(msg, &decoded); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if decoded.Message != payload.Message {
		t.Errorf("message: expected %q, got %q", payload.Message, decoded.Message)
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	_, err := protocol.Decode([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestEncode_ProducesValidJSON(t *testing.T) {
	data, err := protocol.Encode(protocol.TypeAuthPair, protocol.PairPayload{
		PairingToken: "tok",
		AgentID:      "a1",
		Hostname:     "host",
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !json.Valid(data) {
		t.Error("Encode produced invalid JSON")
	}
}

func TestDecodePayload_TypeMismatch(t *testing.T) {
	// Encode a PairPayload but try to decode as ConnectSuccessPayload
	data, err := protocol.Encode(protocol.TypeAuthPair, protocol.PairPayload{
		PairingToken: "tok",
		AgentID:      "a1",
		Hostname:     "host",
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// This should still unmarshal without error (JSON is permissive),
	// but the fields won't match — ConnectSuccessPayload.AgentID will get "a1"
	// from the agentId field in PairPayload.
	var decoded protocol.ConnectSuccessPayload
	if err := protocol.DecodePayload(msg, &decoded); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}
}
