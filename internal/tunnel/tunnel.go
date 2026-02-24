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
	remoteSocket := cfg.RemoteSocketPath
	if remoteSocket == "" {
		// Infer remote socket path by querying the remote user's UID.
		uid, err := m.resolveRemoteUID(ctx, cfg)
		if err != nil {
			return fmt.Errorf("infer remote socket path: %w", err)
		}
		remoteSocket = fmt.Sprintf("/run/user/%s/attn.sock", uid)
		log.Printf("tunnel %q: inferred remote socket path: %s", cfg.Name, remoteSocket)
	}

	// Remove stale socket on the remote before binding.
	m.removeRemoteSocket(ctx, cfg, remoteSocket)

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
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return err
}

// removeRemoteSocket removes a stale socket file on the remote machine.
// Errors are logged but not fatal — the tunnel will fail with a clear
// error if the socket still can't be bound.
func (m *Manager) removeRemoteSocket(ctx context.Context, cfg config.TunnelConfig, socketPath string) {
	args := []string{
		"-o", "ConnectTimeout=10",
	}

	if cfg.IdentityFile != "" {
		idFile := cfg.IdentityFile
		if strings.HasPrefix(idFile, "~/") {
			idFile = expandHome(idFile)
		}
		args = append(args, "-i", idFile)
	}

	dest := cfg.Host
	if cfg.User != "" {
		dest = cfg.User + "@" + cfg.Host
	}
	args = append(args, dest, "rm", "-f", socketPath)

	cmd := exec.CommandContext(ctx, "ssh", args...)
	_ = cmd.Run()
}

// resolveRemoteUID runs a quick SSH command to determine the remote user's UID.
func (m *Manager) resolveRemoteUID(ctx context.Context, cfg config.TunnelConfig) (string, error) {
	args := []string{
		"-o", "ConnectTimeout=10",
	}

	if cfg.IdentityFile != "" {
		idFile := cfg.IdentityFile
		if strings.HasPrefix(idFile, "~/") {
			idFile = expandHome(idFile)
		}
		args = append(args, "-i", idFile)
	}

	dest := cfg.Host
	if cfg.User != "" {
		dest = cfg.User + "@" + cfg.Host
	}
	args = append(args, dest, "id", "-u")

	cmd := exec.CommandContext(ctx, "ssh", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ssh %s id -u: %w", dest, err)
	}

	uid := strings.TrimSpace(string(out))
	if uid == "" {
		return "", fmt.Errorf("ssh %s id -u: empty output", dest)
	}
	return uid, nil
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
