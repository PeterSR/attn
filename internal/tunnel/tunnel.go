package tunnel

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/petersr/attn/internal/config"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 60 * time.Second
)

// Manager manages SSH tunnels for remote socket forwarding.
type Manager struct {
	localSocket string
	tunnels     []config.TunnelConfig
}

// NewManager creates a tunnel manager.
func NewManager(localSocket string, tunnels []config.TunnelConfig) *Manager {
	return &Manager{
		localSocket: localSocket,
		tunnels:     tunnels,
	}
}

// Run starts all configured tunnels. Blocks until ctx is cancelled.
// Each tunnel runs in its own goroutine with auto-reconnect.
func (m *Manager) Run(ctx context.Context) {
	if len(m.tunnels) == 0 {
		return
	}

	for _, t := range m.tunnels {
		go m.runTunnel(ctx, t)
	}

	<-ctx.Done()
}

func (m *Manager) runTunnel(ctx context.Context, cfg config.TunnelConfig) {
	name := cfg.Name
	if name == "" {
		name = cfg.Host
	}

	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("tunnel %q: connecting to %s", name, cfg.Host)

		err := m.execSSH(ctx, cfg)

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			log.Printf("tunnel %q: disconnected: %v", name, err)
		} else {
			log.Printf("tunnel %q: disconnected", name)
		}

		log.Printf("tunnel %q: reconnecting in %v", name, backoff)

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff = backoff * 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (m *Manager) execSSH(ctx context.Context, cfg config.TunnelConfig) error {
	remoteSocket := cfg.RemoteSocket
	if remoteSocket == "" {
		return fmt.Errorf("remote_socket not configured")
	}

	// Build SSH args.
	args := []string{
		"-N", // No remote command
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-R", fmt.Sprintf("%s:%s", remoteSocket, m.localSocket),
	}

	if cfg.IdentityFile != "" {
		idFile := cfg.IdentityFile
		if strings.HasPrefix(idFile, "~/") {
			// Expand ~ manually since exec doesn't use shell.
			if home, err := exec.LookPath("sh"); err == nil {
				_ = home // not needed, just expand ~
			}
			idFile = expandHome(idFile)
		}
		args = append(args, "-i", idFile)
	}

	dest := cfg.Host
	if cfg.User != "" {
		dest = cfg.User + "@" + cfg.Host
	}
	args = append(args, dest)

	cmd := exec.CommandContext(ctx, "ssh", args...)
	return cmd.Run()
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	// Use $HOME since os.UserHomeDir may not be available in all contexts.
	home := ""
	if h, err := exec.Command("sh", "-c", "echo $HOME").Output(); err == nil {
		home = strings.TrimSpace(string(h))
	}
	if home == "" {
		return path
	}
	return home + path[1:]
}
