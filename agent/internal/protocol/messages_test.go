package protocol_test

import (
	"encoding/json"
	"testing"

	"github.com/eyes-on-vps/agent/internal/protocol"
)

func TestEncodeDecode_PairMessage(t *testing.T) {
	payload := protocol.PairPayload{
		PairingToken: "tok-abc123",
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
}

func TestEncodeDecode_ConnectMessage(t *testing.T) {
	payload := protocol.ConnectPayload{
		Token: "jwt-token-here",
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

	if decoded.Token != payload.Token {
		t.Errorf("token: expected %q, got %q", payload.Token, decoded.Token)
	}
}

func TestEncodeDecode_PairSuccessPayload(t *testing.T) {
	payload := protocol.PairSuccessPayload{
		Token: "new-jwt-token",
		Agent: protocol.AgentInfo{
			ID:       "agent-1",
			Hostname: "vps-prod",
			OS:       "linux",
			Arch:     "amd64",
		},
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

	if decoded.Token != payload.Token {
		t.Errorf("token: expected %q, got %q", payload.Token, decoded.Token)
	}
	if decoded.Agent.ID != payload.Agent.ID {
		t.Errorf("agent.id: expected %q, got %q", payload.Agent.ID, decoded.Agent.ID)
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
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !json.Valid(data) {
		t.Error("Encode produced invalid JSON")
	}
}
