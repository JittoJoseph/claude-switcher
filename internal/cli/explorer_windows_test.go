package cli

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A console executable launched from Explorer gets its own console. Cobra
// normally intercepts that launch, even when a command was supplied.
func TestExplorerLaunchRunsCommand(t *testing.T) {
	if os.Getenv("CLAUDE_SWITCHER_EXPLORER_TEST") == "1" {
		os.Args = []string{os.Args[0], "version", "--short"}
		os.Exit(Execute(Build{Version: "explorer-test"}))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestExplorerLaunchRunsCommand$")
	cmd.Env = append(os.Environ(), "CLAUDE_SWITCHER_EXPLORER_TEST=1", "CLAUDE_SWITCHER_NO_UPDATE_CHECK=1")
	// Keep the real new-console behavior without displaying a test window.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010, HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Explorer-style launch: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != "explorer-test" {
		t.Fatalf("expected command output instead of a CLI warning, got %q", got)
	}
}
