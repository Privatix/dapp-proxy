package osconnector

import (
	"os/exec"
	"path/filepath"
)

// changedServicesFileName net services changed to use proxy stored in a file
// with this name in the same directory as the script.
const changedServicesFileName = "netservices"

func changedServicesFileNamePath(script string) string {
	dir, _ := filepath.Split(script)
	return filepath.Join(dir, changedServicesFileName)
}

// ConfigureWithScript uses script to configure proxy.
func ConfigureWithScript(script, host, port string) error {
	chsf := changedServicesFileNamePath(script)

	cmd := exec.Command("/bin/sh", script, "on", chsf, host, port)
	return cmd.Run()
}

// RollbackWithScript uses scrupt to rollback proxy configuration.
func RollbackWithScript(script string) error {
	chsf := changedServicesFileNamePath(script)

	cmd := exec.Command("/bin/sh", script, "off", chsf)
	return cmd.Run()
}