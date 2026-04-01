package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/eyes-on-vps/agent/internal/auth"
	"github.com/eyes-on-vps/agent/internal/protocol"
)

// MessageHandler is called for each non-auth message received from a desktop.
type MessageHandler func(msg protocol.Message)

// LifecycleHandler is called when the first desktop connects or the last disconnects.
// OnActive receives a context canceled when all desktops disconnect, and a broadcast function.
type LifecycleHandler func(ctx context.Context, broadcast func(msgType string, payload any) error)

// Server manages the WebSocket server that desktop clients connect to.
type Server struct {
	auth     *auth.Handler
	handler  MessageHandler
	onActive LifecycleHandler

	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
	// cancel for the active session (first connect → last disconnect)
	activeCancel context.CancelFunc
}

// NewServer creates a transport server.
func NewServer(authHandler *auth.Handler, handler MessageHandler) *Server {
	return &Server{
		auth:    authHandler,
		handler: handler,
		conns:   make(map[*websocket.Conn]struct{}),
	}
}

// OnActive sets a handler called when the first desktop connects.
// The context is canceled when the last desktop disconnects.
func (s *Server) OnActive(handler LifecycleHandler) {
	s.onActive = handler
}

// Run starts the WebSocket server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleConnection(ctx, w, r)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	srv := &http.Server{Handler: mux}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	log.Printf("listening on :%d", port)

	if err := srv.Serve(listener); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleConnection(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("accept: %v", err)
		return
	}
	defer conn.CloseNow()

	if err := s.authenticate(ctx, conn); err != nil {
		log.Printf("auth failed: %v", err)
		return
	}

	s.addConn(ctx, conn)
	defer s.removeConn(conn)

	log.Printf("desktop connected (%d total)", s.connCount())

	for {
		var msg protocol.Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		if s.handler != nil {
			s.handler(msg)
		}
	}
}

func (s *Server) addConn(ctx context.Context, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wasEmpty := len(s.conns) == 0
	s.conns[conn] = struct{}{}

	// First desktop connected — start collectors
	if wasEmpty && s.onActive != nil {
		activeCtx, cancel := context.WithCancel(ctx)
		s.activeCancel = cancel
		go s.onActive(activeCtx, s.Broadcast)
	}
}

func (s *Server) removeConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, conn)
	log.Printf("desktop disconnected (%d remaining)", len(s.conns))

	// Last desktop disconnected — stop collectors
	if len(s.conns) == 0 && s.activeCancel != nil {
		s.activeCancel()
		s.activeCancel = nil
	}
}

func (s *Server) connCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.conns)
}

func (s *Server) authenticate(ctx context.Context, conn *websocket.Conn) error {
	var msg protocol.Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		return fmt.Errorf("read auth: %w", err)
	}

	response, err := s.auth.HandleAuth(msg)
	if err != nil {
		return err
	}

	if err := conn.Write(ctx, websocket.MessageText, response); err != nil {
		return fmt.Errorf("write auth response: %w", err)
	}

	respMsg, err := protocol.Decode(response)
	if err != nil {
		return err
	}

	switch respMsg.Type {
	case protocol.TypeAuthPairSuccess, protocol.TypeAuthConnectSuccess:
		return nil
	default:
		return fmt.Errorf("auth rejected: %s", respMsg.Type)
	}
}

// Broadcast sends a message to all connected desktops.
func (s *Server) Broadcast(msgType string, payload any) error {
	data, err := protocol.Encode(msgType, payload)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	s.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.mu.Unlock()

	for _, c := range conns {
		c.Write(context.Background(), websocket.MessageText, data)
	}
	return nil
}

// Send transmits a message to a specific connection (kept for future use).
func (s *Server) Send(ctx context.Context, conn *websocket.Conn, msgType string, payload any) error {
	data, err := protocol.Encode(msgType, payload)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return conn.Write(ctx, websocket.MessageText, data)
}
