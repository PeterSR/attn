//go:build linux

package dbus

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	godbus "github.com/godbus/dbus/v5"
)

// SessionBus connects to the D-Bus session bus with fallback logic
// for non-interactive contexts (cron, agent hooks, systemd services).
func SessionBus() (*godbus.Conn, error) {
	// Try the standard connection first (uses DBUS_SESSION_BUS_ADDRESS).
	conn, err := godbus.ConnectSessionBus()
	if err == nil {
		return conn, nil
	}

	// Fallback: construct the bus address from XDG_RUNTIME_DIR.
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg == "" {
		xdg = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	busPath := filepath.Join(xdg, "bus")

	if _, statErr := os.Stat(busPath); statErr != nil {
		return nil, fmt.Errorf("D-Bus session bus unavailable (tried standard and %s): %w", busPath, err)
	}

	addr := "unix:path=" + busPath
	conn, err = godbus.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus at %s: %w", addr, err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("D-Bus auth failed: %w", err)
	}

	if err = conn.Hello(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("D-Bus hello failed: %w", err)
	}

	return conn, nil
}

// IsDBusAvailable returns true if a D-Bus session bus socket exists.
func IsDBusAvailable() bool {
	// Check DBUS_SESSION_BUS_ADDRESS first.
	if addr := os.Getenv("DBUS_SESSION_BUS_ADDRESS"); addr != "" {
		return true
	}
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg == "" {
		xdg = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	busPath := filepath.Join(xdg, "bus")
	_, err := net.Dial("unix", busPath)
	return err == nil
}
