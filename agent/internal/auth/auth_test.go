package auth_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/eyes-on-vps/agent/internal/auth"
	"github.com/eyes-on-vps/agent/internal/config"
	"github.com/eyes-on-vps/agent/internal/protocol"
)

func tempConfig(t *testing.T, cfg *config.Config) *config.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	cfg.SetPath(path)
	if err := cfg.Save(); err != nil {
		t.Fatalf("save temp config: %v", err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load temp config: %v", err)
	}
	return loaded
}

func TestHandleAuth_PairSuccess(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		AgentID:       "agent-1",
		Port:          9090,
		SigningSecret: "test-secret-key-for-jwt-signing",
		PairingToken:  "valid-token",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairPayload{PairingToken: "valid-token"})
	msg := protocol.Message{Type: protocol.TypeAuthPair, Payload: raw}

	resp, err := h.HandleAuth(msg)
	if err != nil {
		t.Fatalf("HandleAuth failed: %v", err)
	}

	respMsg, _ := protocol.Decode(resp)
	if respMsg.Type != protocol.TypeAuthPairSuccess {
		t.Errorf("expected %q, got %q", protocol.TypeAuthPairSuccess, respMsg.Type)
	}

	var payload protocol.PairSuccessPayload
	protocol.DecodePayload(respMsg, &payload)

	if payload.Token == "" {
		t.Error("expected non-empty token")
	}
	if payload.Agent.ID != "agent-1" {
		t.Errorf("expected agent ID %q, got %q", "agent-1", payload.Agent.ID)
	}

	// Pairing token should be consumed
	if cfg.PairingToken != "" {
		t.Errorf("pairing token should be cleared, got %q", cfg.PairingToken)
	}
	if cfg.PairedTokenHash == "" {
		t.Error("paired token hash should be set")
	}
}

func TestHandleAuth_PairInvalidToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		AgentID:       "agent-1",
		Port:          9090,
		SigningSecret: "test-secret",
		PairingToken:  "correct-token",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairPayload{PairingToken: "wrong-token"})
	msg := protocol.Message{Type: protocol.TypeAuthPair, Payload: raw}

	resp, err := h.HandleAuth(msg)
	if err != nil {
		t.Fatalf("HandleAuth failed: %v", err)
	}

	respMsg, _ := protocol.Decode(resp)
	if respMsg.Type != protocol.TypeAuthPairError {
		t.Errorf("expected %q, got %q", protocol.TypeAuthPairError, respMsg.Type)
	}
}

func TestHandleAuth_PairNoPairingToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		AgentID:       "agent-1",
		Port:          9090,
		SigningSecret: "test-secret",
		PairingToken:  "", // no token set
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairPayload{PairingToken: "any-token"})
	msg := protocol.Message{Type: protocol.TypeAuthPair, Payload: raw}

	resp, err := h.HandleAuth(msg)
	if err != nil {
		t.Fatalf("HandleAuth failed: %v", err)
	}

	respMsg, _ := protocol.Decode(resp)
	if respMsg.Type != protocol.TypeAuthPairError {
		t.Errorf("expected %q, got %q", protocol.TypeAuthPairError, respMsg.Type)
	}
}

func TestHandleAuth_ConnectSuccess(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		AgentID:       "agent-1",
		Port:          9090,
		SigningSecret: "test-secret-key-for-jwt-signing",
		PairingToken:  "pair-token",
	})
	h := auth.NewHandler(cfg, "test")

	// First pair to get a token
	pairRaw, _ := json.Marshal(protocol.PairPayload{PairingToken: "pair-token"})
	pairMsg := protocol.Message{Type: protocol.TypeAuthPair, Payload: pairRaw}
	pairResp, _ := h.HandleAuth(pairMsg)
	pairRespMsg, _ := protocol.Decode(pairResp)

	var pairPayload protocol.PairSuccessPayload
	protocol.DecodePayload(pairRespMsg, &pairPayload)

	// Now reconnect with the issued token
	connectRaw, _ := json.Marshal(protocol.ConnectPayload{Token: pairPayload.Token})
	connectMsg := protocol.Message{Type: protocol.TypeAuthConnect, Payload: connectRaw}

	resp, err := h.HandleAuth(connectMsg)
	if err != nil {
		t.Fatalf("HandleAuth failed: %v", err)
	}

	respMsg, _ := protocol.Decode(resp)
	if respMsg.Type != protocol.TypeAuthConnectSuccess {
		t.Errorf("expected %q, got %q", protocol.TypeAuthConnectSuccess, respMsg.Type)
	}
}

func TestHandleAuth_ConnectBadToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		AgentID:         "agent-1",
		Port:            9090,
		SigningSecret:   "test-secret",
		PairedTokenHash: "some-hash",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.ConnectPayload{Token: "invalid-jwt"})
	msg := protocol.Message{Type: protocol.TypeAuthConnect, Payload: raw}

	resp, err := h.HandleAuth(msg)
	if err != nil {
		t.Fatalf("HandleAuth failed: %v", err)
	}

	respMsg, _ := protocol.Decode(resp)
	if respMsg.Type != protocol.TypeAuthConnectError {
		t.Errorf("expected %q, got %q", protocol.TypeAuthConnectError, respMsg.Type)
	}
}

func TestIsAuthMessage(t *testing.T) {
	tests := []struct {
		msgType  string
		expected bool
	}{
		{protocol.TypeAuthPair, true},
		{protocol.TypeAuthConnect, true},
		{protocol.TypeAuthPairSuccess, false},
		{protocol.TypeAuthPairError, false},
		{"metrics:system", false},
	}

	for _, tt := range tests {
		if got := auth.IsAuthMessage(tt.msgType); got != tt.expected {
			t.Errorf("IsAuthMessage(%q) = %v, want %v", tt.msgType, got, tt.expected)
		}
	}
}
