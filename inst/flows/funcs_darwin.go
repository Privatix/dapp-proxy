package flows

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/privatix/dapp-proxy/plugin/adapter"
	"github.com/privatix/dapp-proxy/plugin/osconnector"
)

func removeOSProxyConfigurationIfAny(p *ProxyInstallation) error {
	if p.IsAgent {
		return nil
	}
	err := osconnector.RollbackWithScript(p.configureProxyScript())
	if err != osconnector.ErrRollbackNotNeeded {
		return err
	}
	return nil
}

func configureOSFirewall(p *ProxyInstallation) error {
	// Need to open firewall for incoming traffic only on agents side.
	if !p.IsAgent {
		return nil
	}

	config := new(adapter.Config)
	readJSON(p.pluginAgentConfigPath(), config)
	cmd := exec.Command(p.osxFilrewallScript(), "on", fmt.Sprint(config.V2Ray.InboundPort), p.osxFirewallRuleFile())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not execute configure firewall script: %v", err)
	}
	return nil
}

func readAgentInboundPort(p *ProxyInstallation) (uint, error) {
	text, err := ioutil.ReadFile(p.pluginAgentConfigPath())
	if err != nil {
		return 0, fmt.Errorf("could not read agent config: %v", err)
	}
	var conf adapter.Config
	if err := json.Unmarshal(text, &conf); err != nil {
		return 0, fmt.Errorf("could not parse agent config: %v", err)
	}

	return conf.V2Ray.InboundPort, nil
}

func rollbackOSConfiguration(p *ProxyInstallation) error {
	if !p.IsAgent {
		return nil
	}
	cmd := exec.Command(p.osxFilrewallScript(), "off")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not execute configure firewall script: %v", err)
	}
	return nil
}
