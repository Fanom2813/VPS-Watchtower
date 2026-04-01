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

// MessageHandler is called for each non-auth message received from the desktop.
type MessageHandler func(msg protocol.Message)

// Server manages the WebSocket server that desktop clients connect to.
type Server struct {
	auth    *auth.Handler
	handler MessageHandler

	mu   sync.Mutex
	conn *websocket.Conn
}

// NewServer creates a transport server.
func NewServer(authHandler *auth.Handler, handler MessageHandler) *Server {
	return &Server{
		auth:    authHandler,
		handler: handler,
	}
}

// Run starts the WebSocket server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleConnection(r.Context(), w, r)
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

	// Authenticate the desktop
	if err := s.authenticate(ctx, conn); err != nil {
		log.Printf("auth failed: %v", err)
		return
	}

	s.mu.Lock()
	// Close any existing connection (only one desktop at a time)
	if s.conn != nil {
		s.conn.Close(websocket.StatusGoingAway, "replaced by new connection")
	}
	s.conn = conn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
		}
		s.mu.Unlock()
	}()

	log.Println("desktop connected and authenticated")

	// Listen for messages
	for {
		var msg protocol.Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			log.Printf("desktop disconnected: %v", err)
			return
		}
		if s.handler != nil {
			s.handler(msg)
		}
	}
}

func (s *Server) authenticate(ctx context.Context, conn *websocket.Conn) error {
	// Read auth message from desktop
	var msg protocol.Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		return fmt.Errorf("read auth: %w", err)
	}

	// Delegate to auth handler
	response, err := s.auth.HandleAuth(msg)
	if err != nil {
		return err
	}

	// Send response back
	if err := conn.Write(ctx, websocket.MessageText, response); err != nil {
		return fmt.Errorf("write auth response: %w", err)
	}

	// Check if it was a success or error response
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

// Send transmits a message to the connected desktop.
func (s *Server) Send(ctx context.Context, msgType string, payload any) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("no desktop connected")
	}

	data, err := protocol.Encode(msgType, payload)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	return conn.Write(ctx, websocket.MessageText, data)
}
