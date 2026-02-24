package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/notification"
	"github.com/petersr/attn/internal/relay"
	"github.com/petersr/attn/internal/tunnel"
)

// ServeCmd starts the relay server.
type ServeCmd struct {
	Socket    string `help:"Unix socket path." default:"" env:"ATTN_SOCK"`
	Install   bool   `help:"Install and enable systemd user service."`
	Uninstall bool   `help:"Disable and remove systemd user service."`
}

const systemdServiceTemplate = `[Unit]
Description=attn notification relay
After=graphical-session.target

[Service]
ExecStart={{.ExecStart}} serve --socket {{.Socket}}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

func (s *ServeCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
	}

	socketPath := s.Socket
	if socketPath == "" {
		socketPath = cfg.Serve.SocketPath
	}
	if socketPath == "" {
		socketPath = relay.DefaultSocketPath()
	}

	if s.Install {
		return installService(socketPath)
	}
	if s.Uninstall {
		return uninstallService()
	}

	return runServer(cfg, socketPath)
}

func runServer(cfg config.Config, socketPath string) error {
	srv := &relay.Server{
		SocketPath: socketPath,
		DispatchFunc: func(ctx context.Context, n notification.Notification, hops int) error {
			entries := buildChannelEntries(cfg, hops)
			state := channel.DetectScreenState(entries, hops)
			return channel.DispatchFiltered(ctx, entries, state, n)
		},
	}

	if err := srv.Listen(); err != nil {
		return err
	}
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start SSH tunnels if configured.
	if len(cfg.Serve.Tunnels) > 0 {
		tm := tunnel.NewManager(socketPath, cfg.Serve.Tunnels)
		go tm.Run(ctx)
	}

	// Handle shutdown signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		cancel()
	}()

	return srv.Serve(ctx)
}

func installService(socketPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	servicePath := filepath.Join(serviceDir, "attn.service")

	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("create systemd user dir: %w", err)
	}

	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("create service file: %w", err)
	}
	defer f.Close()

	data := struct {
		ExecStart string
		Socket    string
	}{
		ExecStart: exe,
		Socket:    socketPath,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("write service file: %w", err)
	}

	fmt.Printf("wrote %s\n", servicePath)

	// Enable and start the service.
	cmds := []struct {
		desc string
		args []string
	}{
		{"reload systemd", []string{"systemctl", "--user", "daemon-reload"}},
		{"enable service", []string{"systemctl", "--user", "enable", "attn.service"}},
		{"start service", []string{"systemctl", "--user", "start", "attn.service"}},
	}

	for _, c := range cmds {
		cmd := newCommand(c.args[0], c.args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %w: %s", c.desc, err, string(out))
		}
		fmt.Printf("%s: ok\n", c.desc)
	}

	fmt.Println("\nattn relay service installed and running.")
	fmt.Println("Check status: systemctl --user status attn")
	return nil
}

func uninstallService() error {
	cmds := []struct {
		desc string
		args []string
	}{
		{"stop service", []string{"systemctl", "--user", "stop", "attn.service"}},
		{"disable service", []string{"systemctl", "--user", "disable", "attn.service"}},
	}

	for _, c := range cmds {
		cmd := newCommand(c.args[0], c.args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v: %s\n", c.desc, err, string(out))
			// Continue anyway — service may already be stopped.
		} else {
			fmt.Printf("%s: ok\n", c.desc)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	servicePath := filepath.Join(home, ".config", "systemd", "user", "attn.service")
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove service file: %w", err)
	}
	fmt.Printf("removed %s\n", servicePath)

	_ = newCommand("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("\nattn relay service uninstalled.")
	return nil
}
