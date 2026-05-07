package rewrite

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func TryRtkRewrite(cmd string) (string, bool) {
	if cmd == "" {
		return "", false
	}

	path, err := exec.LookPath("rtk")
	if err != nil {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "rewrite", cmd).Output()
	if err != nil {
		return "", false
	}

	rewritten := strings.TrimSpace(string(out))
	if rewritten == "" || rewritten == cmd {
		return "", false
	}

	return rewritten, true
}
