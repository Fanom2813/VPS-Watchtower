package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/eyes-on-vps/agent/internal/config"
	"github.com/eyes-on-vps/agent/internal/protocol"
	"github.com/eyes-on-vps/agent/internal/sysinfo"
)

// Handler manages server-side authentication for incoming desktop connections.
type Handler struct {
	cfg     *config.Config
	sysInfo sysinfo.StaticInfo
	version string
}

// NewHandler creates an auth handler backed by the given config.
func NewHandler(cfg *config.Config, version string) *Handler {
	return &Handler{
		cfg:     cfg,
		sysInfo: sysinfo.Collect(),
		version: version,
	}
}

// HandleAuth processes an incoming auth message from a desktop client.
// Returns the response bytes to send back. If error is non-nil, the
// connection should be closed (internal failure, not auth rejection).
func (h *Handler) HandleAuth(msg protocol.Message) ([]byte, error) {
	switch msg.Type {
	case protocol.TypeAuthPair:
		return h.handlePair(msg)
	case protocol.TypeAuthConnect:
		return h.handleConnect(msg)
	default:
		return protocol.Encode(protocol.TypeAuthPairError, protocol.ErrorPayload{
			Message: "unknown auth message type",
		})
	}
}

func (h *Handler) handlePair(msg protocol.Message) ([]byte, error) {
	var payload protocol.PairPayload
	if err := protocol.DecodePayload(msg, &payload); err != nil {
		return nil, fmt.Errorf("decode pair payload: %w", err)
	}

	if h.cfg.PairingToken == "" {
		return protocol.Encode(protocol.TypeAuthPairError, protocol.ErrorPayload{
			Message: "agent is not accepting pairing requests",
		})
	}

	if payload.PairingToken != h.cfg.PairingToken {
		return protocol.Encode(protocol.TypeAuthPairError, protocol.ErrorPayload{
			Message: "invalid pairing token",
		})
	}

	// Issue JWT for the desktop
	token, err := h.signToken()
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	// Persist: store token hash, consume pairing token
	h.cfg.PairedTokenHash = hashToken(token)
	h.cfg.PairingToken = ""
	if err := h.cfg.Save(); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	return protocol.Encode(protocol.TypeAuthPairSuccess, protocol.PairSuccessPayload{
		Token: token,
		Agent: h.agentInfo(),
	})
}

func (h *Handler) handleConnect(msg protocol.Message) ([]byte, error) {
	var payload protocol.ConnectPayload
	if err := protocol.DecodePayload(msg, &payload); err != nil {
		return nil, fmt.Errorf("decode connect payload: %w", err)
	}

	// Verify JWT signature
	if err := h.verifyToken(payload.Token); err != nil {
		return protocol.Encode(protocol.TypeAuthConnectError, protocol.ErrorPayload{
			Message: "invalid or expired token",
		})
	}

	// Verify token hash matches the paired desktop
	if h.cfg.PairedTokenHash == "" || hashToken(payload.Token) != h.cfg.PairedTokenHash {
		return protocol.Encode(protocol.TypeAuthConnectError, protocol.ErrorPayload{
			Message: "token revoked",
		})
	}

	return protocol.Encode(protocol.TypeAuthConnectSuccess, protocol.ConnectSuccessPayload{
		Agent: h.agentInfo(),
	})
}

func (h *Handler) agentInfo() protocol.AgentInfo {
	return protocol.AgentInfo{
		ID:       h.cfg.AgentID,
		Hostname: h.sysInfo.Hostname,
		OS:       h.sysInfo.OS,
		Arch:     h.sysInfo.Arch,
		Distro:   h.sysInfo.Distro,
		Version:  h.version,
	}
}

func (h *Handler) signToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": h.cfg.AgentID,
	})
	return token.SignedString([]byte(h.cfg.SigningSecret))
}

func (h *Handler) verifyToken(tokenString string) error {
	_, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.cfg.SigningSecret), nil
	})
	return err
}

// IsAuthMessage returns true if the message type is an auth request.
func IsAuthMessage(msgType string) bool {
	switch msgType {
	case protocol.TypeAuthPair, protocol.TypeAuthConnect:
		return true
	}
	return false
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
