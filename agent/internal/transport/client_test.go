package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/eyes-on-vps/agent/internal/auth"
	"github.com/eyes-on-vps/agent/internal/config"
	"github.com/eyes-on-vps/agent/internal/protocol"
	"github.com/eyes-on-vps/agent/internal/transport"
)

// mockServer simulates the desktop WebSocket server for testing.
type mockServer struct {
	server    *httptest.Server
	onConnect func(conn *websocket.Conn)
}

func newMockServer(t *testing.T, handler func(conn *websocket.Conn)) *mockServer {
	t.Helper()
	ms := &mockServer{onConnect: handler}
	ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("mock server accept error: %v", err)
			return
		}
		defer conn.CloseNow()
		ms.onConnect(conn)
	}))
	t.Cleanup(ms.server.Close)
	return ms
}

func (ms *mockServer) wsURL() string {
	return "ws" + strings.TrimPrefix(ms.server.URL, "http")
}

func TestConnect_PairingFlow(t *testing.T) {
	pairDone := make(chan struct{})

	ms := newMockServer(t, func(conn *websocket.Conn) {
		ctx := context.Background()

		var msg protocol.Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Logf("server read: %v", err)
			return
		}

		// After successful pairing, reconnects will send auth:connect — that's correct.
		if msg.Type == protocol.TypeAuthConnect {
			resp := protocol.Message{Type: protocol.TypeAuthConnectSuccess}
			resp.Payload, _ = json.Marshal(protocol.ConnectSuccessPayload{AgentID: "agent-test"})
			wsjson.Write(ctx, conn, resp)
			time.Sleep(100 * time.Millisecond)
			conn.Close(websocket.StatusNormalClosure, "done")
			return
		}

		if msg.Type != protocol.TypeAuthPair {
			t.Errorf("expected %q, got %q", protocol.TypeAuthPair, msg.Type)
			return
		}

		var payload protocol.PairPayload
		if err := protocol.DecodePayload(msg, &payload); err != nil {
			t.Errorf("decode payload: %v", err)
			return
		}

		if payload.PairingToken != "test-pair-token" {
			t.Errorf("expected pairing token %q, got %q", "test-pair-token", payload.PairingToken)
			return
		}

		resp := protocol.Message{Type: protocol.TypeAuthPairSuccess}
		resp.Payload, _ = json.Marshal(protocol.PairSuccessPayload{AgentToken: "issued-jwt"})
		if err := wsjson.Write(ctx, conn, resp); err != nil {
			t.Errorf("server write: %v", err)
			return
		}

		close(pairDone)
		time.Sleep(100 * time.Millisecond)
		conn.Close(websocket.StatusNormalClosure, "done")
	})

	cfg := &config.Config{
		ServerURL:    ms.wsURL(),
		PairingToken: "test-pair-token",
		AgentID:      "agent-test",
	}
	cfg.SetPath(t.TempDir() + "/config.json")
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	handler := auth.NewHandler(cfg, "test")
	client := transport.NewClient(cfg, handler, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		select {
		case <-pairDone:
			// Give time for the client to process the response.
			time.Sleep(200 * time.Millisecond)
			cancel()
		case <-ctx.Done():
		}
	}()

	_ = client.Run(ctx)

	if cfg.AgentToken != "issued-jwt" {
		t.Errorf("expected agentToken %q, got %q", "issued-jwt", cfg.AgentToken)
	}
	if cfg.PairingToken != "" {
		t.Errorf("expected pairingToken cleared, got %q", cfg.PairingToken)
	}
}

func TestConnect_ReconnectFlow(t *testing.T) {
	ms := newMockServer(t, func(conn *websocket.Conn) {
		ctx := context.Background()

		var msg protocol.Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Errorf("server read: %v", err)
			return
		}

		if msg.Type != protocol.TypeAuthConnect {
			t.Errorf("expected %q, got %q", protocol.TypeAuthConnect, msg.Type)
			return
		}

		var payload protocol.ConnectPayload
		if err := protocol.DecodePayload(msg, &payload); err != nil {
			t.Errorf("decode payload: %v", err)
			return
		}

		if payload.AgentToken != "existing-jwt" {
			t.Errorf("expected agent token %q, got %q", "existing-jwt", payload.AgentToken)
		}

		resp := protocol.Message{Type: protocol.TypeAuthConnectSuccess}
		resp.Payload, _ = json.Marshal(protocol.ConnectSuccessPayload{AgentID: "agent-test"})
		if err := wsjson.Write(ctx, conn, resp); err != nil {
			t.Errorf("server write: %v", err)
			return
		}

		time.Sleep(100 * time.Millisecond)
		conn.Close(websocket.StatusNormalClosure, "done")
	})

	cfg := &config.Config{
		ServerURL:  ms.wsURL(),
		AgentToken: "existing-jwt",
		AgentID:    "agent-test",
	}
	cfg.SetPath(t.TempDir() + "/config.json")
	cfg.Save()

	handler := auth.NewHandler(cfg, "test")
	client := transport.NewClient(cfg, handler, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = client.Run(ctx)
}

func TestConnect_AuthError(t *testing.T) {
	ms := newMockServer(t, func(conn *websocket.Conn) {
		ctx := context.Background()

		var msg protocol.Message
		wsjson.Read(ctx, conn, &msg)

		resp := protocol.Message{Type: protocol.TypeAuthPairError}
		resp.Payload, _ = json.Marshal(protocol.ErrorPayload{Message: "bad token"})
		wsjson.Write(ctx, conn, resp)

		time.Sleep(50 * time.Millisecond)
		conn.Close(websocket.StatusNormalClosure, "")
	})

	cfg := &config.Config{
		ServerURL:    ms.wsURL(),
		PairingToken: "wrong-token",
		AgentID:      "agent-test",
	}
	cfg.SetPath(t.TempDir() + "/config.json")
	cfg.Save()

	handler := auth.NewHandler(cfg, "test")
	client := transport.NewClient(cfg, handler, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run will keep retrying — we just verify it doesn't panic and respects context.
	_ = client.Run(ctx)

	// Agent token should NOT be set after auth failure.
	if cfg.AgentToken != "" {
		t.Errorf("agentToken should be empty after auth failure, got %q", cfg.AgentToken)
	}
}

func TestConnect_MessageHandler(t *testing.T) {
	var received []protocol.Message
	var mu sync.Mutex

	ms := newMockServer(t, func(conn *websocket.Conn) {
		ctx := context.Background()

		// Read and respond to auth.
		var msg protocol.Message
		wsjson.Read(ctx, conn, &msg)

		resp := protocol.Message{Type: protocol.TypeAuthConnectSuccess}
		resp.Payload, _ = json.Marshal(protocol.ConnectSuccessPayload{AgentID: "agent-test"})
		wsjson.Write(ctx, conn, resp)

		// Send a data message.
		dataPayload, _ := json.Marshal(map[string]string{"cpu": "42%"})
		dataMsg := protocol.Message{Type: "metrics:update", Payload: dataPayload}
		wsjson.Write(ctx, conn, dataMsg)

		time.Sleep(200 * time.Millisecond)
		conn.Close(websocket.StatusNormalClosure, "done")
	})

	cfg := &config.Config{
		ServerURL:  ms.wsURL(),
		AgentToken: "valid-jwt",
		AgentID:    "agent-test",
	}
	cfg.SetPath(t.TempDir() + "/config.json")
	cfg.Save()

	handler := auth.NewHandler(cfg, "test")
	client := transport.NewClient(cfg, handler, func(msg protocol.Message) {
		mu.Lock()
		received = append(received, msg)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = client.Run(ctx)

	mu.Lock()
	defer mu.Unlock()

	if len(received) == 0 {
		t.Fatal("expected at least 1 message from handler")
	}

	if received[0].Type != "metrics:update" {
		t.Errorf("expected message type %q, got %q", "metrics:update", received[0].Type)
	}
}

func TestConnect_ServerDown(t *testing.T) {
	cfg := &config.Config{
		ServerURL:  "ws://127.0.0.1:19999",
		AgentToken: "some-jwt",
		AgentID:    "agent-test",
	}
	cfg.SetPath(t.TempDir() + "/config.json")
	cfg.Save()

	handler := auth.NewHandler(cfg, "test")
	client := transport.NewClient(cfg, handler, nil)

	// Should respect context cancellation even when server is unreachable.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Run(ctx)
	if err == nil {
		t.Error("expected error when context expires")
	}
}
