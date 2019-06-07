package osconnector

import (
	"fmt"
	"os"
	"os/exec"
)

// ConfigureWithScript uses script to configure proxy.
func ConfigureWithScript(script, backupFile string, port int) error {
	cmd := exec.Command("/bin/sh", script, "on", backupFile, "127.0.0.1", fmt.Sprint(port))
	return cmd.Run()
}

// RollbackWithScript uses scrupt to rollback proxy configuration.
func RollbackWithScript(script, backupFile string) error {
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return ErrRollbackNotNeeded
	}

	cmd := exec.Command("/bin/sh", script, "off", backupFile)
	return cmd.Run()
}
