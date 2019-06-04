package osconnector

import (
	"fmt"
	"path/filepath"

	"github.com/privatix/dapp-proxy/winutils"
)

const previousProxyBackupFile = "previousProxy.json"

func backupFilePath(script string) string {
	dir, _ := filepath.Split(script)
	return filepath.Join(dir, previousProxyBackupFile)
}

// ConfigureWithScript uses script to configure proxy.
func ConfigureWithScript(script string, port int) error {
	backupFile := backupFilePath(script)
	return winutils.RunPowershellScript(script, "-Action", "set",
		"-ProxyOffSettingsPath", backupFile, "-LocalSocksPort", fmt.Sprint(port))
}

// RollbackWithScript uses scrupt to rollback proxy configuration.
func RollbackWithScript(script string) error {
	backupFile := backupFilePath(script)
	return winutils.RunPowershellScript(script, "-Action", "restore",
		"-ProxyOffSettingsPath", backupFile)
}
