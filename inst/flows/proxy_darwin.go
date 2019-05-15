package flows

func (p *ProxyInstallation) configureProxyScript() string {
	return p.prodPathJoin("data/scripts/configuresocksfirewallproxy.sh")
}
