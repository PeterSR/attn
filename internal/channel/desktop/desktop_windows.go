// EXPERIMENTAL: Windows support is untested. Contributions welcome.

//go:build windows

package desktop

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends desktop notifications via PowerShell on Windows.
// Attempts BurntToast module first, falls back to basic .NET toast.
type Channel struct{}

func New() *Channel {
	return &Channel{}
}

func (c *Channel) Name() string { return "desktop" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	body := n.FormatBody()

	// Escape single quotes for PowerShell strings.
	escBody := strings.ReplaceAll(body, "'", "''")
	escTitle := strings.ReplaceAll(n.Title, "'", "''")

	// Try BurntToast first (richer notifications).
	btScript := fmt.Sprintf(
		`New-BurntToastNotification -Text '%s', '%s'`,
		escTitle, escBody,
	)
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", btScript)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: basic balloon tip via .NET.
	fallback := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$n = New-Object System.Windows.Forms.NotifyIcon
$n.Icon = [System.Drawing.SystemIcons]::Information
$n.BalloonTipTitle = '%s'
$n.BalloonTipText = '%s'
$n.Visible = $true
$n.ShowBalloonTip(5000)
Start-Sleep -Seconds 1
$n.Dispose()
`, escTitle, escBody)

	cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", fallback)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("powershell notification: %w: %s", err, string(out))
	}
	return nil
}
