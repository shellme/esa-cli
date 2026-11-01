package mac

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SendNotification sends a macOS notification
func SendNotification(title, message string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	script := fmt.Sprintf(`
		display notification "%s" with title "%s" sound name "Glass"
	`, message, title)

	return exec.Command("osascript", "-e", script).Run()
}
