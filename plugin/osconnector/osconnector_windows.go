package osconnector

import (
	"fmt"

	"github.com/privatix/dapp-proxy/winutils"
)

// ConfigureWithScript uses script to configure proxy.
func ConfigureWithScript(script, backupFile string, port int) error {
	return winutils.RunPowershellScript(script, "-Action", "set",
		"-ProxyOffSettingsPath", backupFile, "-LocalSocksPort", fmt.Sprint(port))
}

// RollbackWithScript uses scrupt to rollback proxy configuration.
func RollbackWithScript(script, backupFile string) error {
	return winutils.RunPowershellScript(script, "-Action", "restore",
		"-ProxyOffSettingsPath", backupFile)
}
