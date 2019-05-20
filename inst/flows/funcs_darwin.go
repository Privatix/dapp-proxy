package flows

import "github.com/privatix/dapp-proxy/plugin/osconnector"

func removeOSProxyConfigurationIfAny(p *ProxyInstallation) error {
	err := osconnector.RollbackWithScript(p.configureProxyScript())
	if err != osconnector.ErrRollbackNotNeeded {
		return err
	}
	return nil
}
