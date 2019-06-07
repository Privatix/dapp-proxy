package osconnector

import (
	"fmt"
	"os/exec"
)

// ConfigureWithScript uses script to configure proxy.
func ConfigureWithScript(script, backupFile string, port int) error {
	return runScript(script, "on", backupFile, fmt.Sprint(port))
}

// RollbackWithScript uses scrupt to rollback proxy configuration.
func RollbackWithScript(script, backupFile string) error {
	return runScript(script, "off", backupFile)
}

func runScript(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not change proxy settings: %v", err)
	}
	return nil
}
