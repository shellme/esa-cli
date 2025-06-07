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

// OpenInFinder opens the given path in Finder
func OpenInFinder(path string) error {
	return exec.Command("open", path).Run()
}

// OpenInEditor opens the given file in the default editor
func OpenInEditor(filename string) error {
	return exec.Command("open", filename).Run()
}

// AddSpotlightMetadata adds metadata to the file for Spotlight search
func AddSpotlightMetadata(filename string, title, category string) error {
	// Add title metadata
	if err := exec.Command("xattr", "-w", "com.apple.metadata:kMDItemTitle", title, filename).Run(); err != nil {
		return err
	}

	// Add category metadata
	if err := exec.Command("xattr", "-w", "com.apple.metadata:kMDItemDescription", category, filename).Run(); err != nil {
		return err
	}

	return nil
}
