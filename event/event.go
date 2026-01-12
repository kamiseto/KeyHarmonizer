package event

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// Eve_Dialog shows a simple alert dialog on each supported OS.
// robotgo v1 removed ShowAlert, so we shell out to native tools instead.
func Eve_Dialog(target string, dialog string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", fmt.Sprintf(`display alert %q message %q`, target, dialog))
	case "windows":
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`[System.Windows.MessageBox]::Show(%q,%q)`, dialog, target))
	case "linux":
		cmd = exec.Command("zenity", "--info", "--title", target, "--text", dialog)
	default:
		return
	}

	if err := cmd.Run(); err != nil {
		log.Printf("failed to show dialog: %v", err)
	}
}
