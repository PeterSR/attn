package relay

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/notification"
)

// Server listens on a Unix socket and dispatches incoming notifications.
type Server struct {
	SocketPath string
	Channels   []channel.Channel
	listener   net.Listener
}

// DefaultSocketPath returns the default socket path.
func DefaultSocketPath() string {
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg == "" {
		xdg = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	return filepath.Join(xdg, "attn.sock")
}

// Listen starts the server. Call Serve() after this.
func (s *Server) Listen() error {
	// Remove stale socket if it exists.
	if _, err := os.Stat(s.SocketPath); err == nil {
		// Check if something is already listening.
		conn, err := net.Dial("unix", s.SocketPath)
		if err == nil {
			conn.Close()
			return fmt.Errorf("another instance is already listening on %s", s.SocketPath)
		}
		os.Remove(s.SocketPath)
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(s.SocketPath), 0700); err != nil {
		return fmt.Errorf("create socket directory: %w", err)
	}

	l, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.SocketPath, err)
	}
	if err := os.Chmod(s.SocketPath, 0600); err != nil {
		l.Close()
		return fmt.Errorf("chmod socket: %w", err)
	}
	s.listener = l
	return nil
}

// Serve accepts connections until the context is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	log.Printf("relay: listening on %s", s.SocketPath)

	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("relay: accept error: %v", err)
				continue
			}
		}
		go s.handleConn(ctx, conn)
	}
}

// Close cleans up the socket.
func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.SocketPath)
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	enc := json.NewEncoder(conn)

	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			_ = enc.Encode(Response{OK: false, Error: "invalid JSON"})
			continue
		}

		switch msg.Type {
		case "ping":
			_ = enc.Encode(Response{OK: true})

		case "notify":
			if msg.Notify == nil {
				_ = enc.Encode(Response{OK: false, Error: "missing notify payload"})
				continue
			}
			s.dispatch(ctx, *msg.Notify)
			_ = enc.Encode(Response{OK: true})

		default:
			_ = enc.Encode(Response{OK: false, Error: "unknown message type"})
		}
	}
}

func (s *Server) dispatch(ctx context.Context, n notification.Notification) {
	if err := channel.Dispatch(ctx, s.Channels, n); err != nil {
		log.Printf("relay: dispatch error: %v", err)
	}
}
