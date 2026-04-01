package auth_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

func TestNeedsPairing_NoToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "pair-123",
		AgentID:      "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	if !h.NeedsPairing() {
		t.Error("expected NeedsPairing=true when no agent token")
	}
}

func TestNeedsPairing_HasToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:  "ws://localhost:9000",
		AgentToken: "some-jwt",
		AgentID:    "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	if h.NeedsPairing() {
		t.Error("expected NeedsPairing=false when agent token exists")
	}
}

func TestBuildAuthMessage_Pairing(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "pair-token-abc",
		AgentID:      "agent-42",
	})
	h := auth.NewHandler(cfg, "test")

	data, err := h.BuildAuthMessage()
	if err != nil {
		t.Fatalf("BuildAuthMessage failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthPair {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthPair, msg.Type)
	}

	var payload protocol.PairPayload
	if err := protocol.DecodePayload(msg, &payload); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if payload.PairingToken != "pair-token-abc" {
		t.Errorf("pairingToken: expected %q, got %q", "pair-token-abc", payload.PairingToken)
	}
	if payload.AgentID != "agent-42" {
		t.Errorf("agentId: expected %q, got %q", "agent-42", payload.AgentID)
	}
	if payload.Hostname == "" {
		t.Error("hostname should not be empty")
	}
}

func TestBuildAuthMessage_Connect(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:  "ws://localhost:9000",
		AgentToken: "jwt-reconnect-token",
		AgentID:    "agent-42",
	})
	h := auth.NewHandler(cfg, "test")

	data, err := h.BuildAuthMessage()
	if err != nil {
		t.Fatalf("BuildAuthMessage failed: %v", err)
	}

	msg, err := protocol.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != protocol.TypeAuthConnect {
		t.Errorf("expected type %q, got %q", protocol.TypeAuthConnect, msg.Type)
	}

	var payload protocol.ConnectPayload
	if err := protocol.DecodePayload(msg, &payload); err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if payload.AgentToken != "jwt-reconnect-token" {
		t.Errorf("agentToken: expected %q, got %q", "jwt-reconnect-token", payload.AgentToken)
	}
}

func TestHandleResponse_PairSuccess(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "pair-123",
		AgentID:      "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairSuccessPayload{AgentToken: "new-jwt"})
	msg := protocol.Message{Type: protocol.TypeAuthPairSuccess, Payload: raw}

	if err := h.HandleResponse(msg); err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}

	// Verify config was updated and persisted
	reloaded, err := config.Load(filepath.Join(t.TempDir()))
	// Instead, verify in-memory state
	_ = reloaded
	_ = err

	if h.NeedsPairing() {
		t.Error("should not need pairing after successful pair")
	}
}

func TestHandleResponse_PairSuccess_PersistsToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "pair-123",
		AgentID:      "agent-1",
	}
	cfg.SetPath(path)
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairSuccessPayload{AgentToken: "persisted-jwt"})
	msg := protocol.Message{Type: protocol.TypeAuthPairSuccess, Payload: raw}

	if err := h.HandleResponse(msg); err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}

	// Reload from disk and verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if saved.AgentToken != "persisted-jwt" {
		t.Errorf("saved agentToken: expected %q, got %q", "persisted-jwt", saved.AgentToken)
	}
	if saved.PairingToken != "" {
		t.Errorf("saved pairingToken should be cleared, got %q", saved.PairingToken)
	}
}

func TestHandleResponse_ConnectSuccess(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:  "ws://localhost:9000",
		AgentToken: "existing-jwt",
		AgentID:    "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.ConnectSuccessPayload{AgentID: "agent-1"})
	msg := protocol.Message{Type: protocol.TypeAuthConnectSuccess, Payload: raw}

	if err := h.HandleResponse(msg); err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}
}

func TestHandleResponse_PairError(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "bad-token",
		AgentID:      "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.ErrorPayload{Message: "invalid token"})
	msg := protocol.Message{Type: protocol.TypeAuthPairError, Payload: raw}

	err := h.HandleResponse(msg)
	if err == nil {
		t.Fatal("expected error for pair failure")
	}
	if err.Error() != "pairing failed: invalid token" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleResponse_ConnectError(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:  "ws://localhost:9000",
		AgentToken: "expired-jwt",
		AgentID:    "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.ErrorPayload{Message: "token expired"})
	msg := protocol.Message{Type: protocol.TypeAuthConnectError, Payload: raw}

	err := h.HandleResponse(msg)
	if err == nil {
		t.Fatal("expected error for connect failure")
	}
	if err.Error() != "authentication failed: token expired" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleResponse_UnexpectedType(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL: "ws://localhost:9000",
		AgentID:   "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	msg := protocol.Message{Type: "unknown:type", Payload: nil}
	err := h.HandleResponse(msg)
	if err == nil {
		t.Fatal("expected error for unexpected type")
	}
}

func TestHandleResponse_EmptyAgentToken(t *testing.T) {
	cfg := tempConfig(t, &config.Config{
		ServerURL:    "ws://localhost:9000",
		PairingToken: "pair-123",
		AgentID:      "agent-1",
	})
	h := auth.NewHandler(cfg, "test")

	raw, _ := json.Marshal(protocol.PairSuccessPayload{AgentToken: ""})
	msg := protocol.Message{Type: protocol.TypeAuthPairSuccess, Payload: raw}

	err := h.HandleResponse(msg)
	if err == nil {
		t.Fatal("expected error for empty agent token")
	}
}

func TestIsAuthResponse(t *testing.T) {
	tests := []struct {
		msgType  string
		expected bool
	}{
		{protocol.TypeAuthPairSuccess, true},
		{protocol.TypeAuthPairError, true},
		{protocol.TypeAuthConnectSuccess, true},
		{protocol.TypeAuthConnectError, true},
		{protocol.TypeAuthPair, false},
		{protocol.TypeAuthConnect, false},
		{"metrics:data", false},
	}

	for _, tt := range tests {
		if got := auth.IsAuthResponse(tt.msgType); got != tt.expected {
			t.Errorf("IsAuthResponse(%q) = %v, want %v", tt.msgType, got, tt.expected)
		}
	}
}

func makeJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return s
}

func TestValidateToken_Valid(t *testing.T) {
	tok := makeJWT(t, jwt.MapClaims{
		"sub": "agent-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if err := auth.ValidateToken(tok); err != nil {
		t.Errorf("expected valid token, got: %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	tok := makeJWT(t, jwt.MapClaims{
		"sub": "agent-1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	err := auth.ValidateToken(tok)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateToken_NoExpiration(t *testing.T) {
	tok := makeJWT(t, jwt.MapClaims{
		"sub": "agent-1",
	})
	if err := auth.ValidateToken(tok); err != nil {
		t.Errorf("token without exp should be valid, got: %v", err)
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	err := auth.ValidateToken("not-a-jwt")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}
