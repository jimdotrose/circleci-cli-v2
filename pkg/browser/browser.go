package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open launches url in the user's default web browser.
// It shells out to the platform-appropriate opener command and returns any error.
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux and other unixes
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not open browser: %w", err)
	}
	return nil
}
