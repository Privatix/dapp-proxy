package flows

func removeOSProxyConfigurationIfAny(p *ProxyInstallation) error {
	if p.IsAgent {
		return nil
	}
	return nil
}

func configureOSFirewall(p *ProxyInstallation) error {
	return nil
}

func rollbackOSConfiguration(p *ProxyInstallation) error {
	return nil
}
