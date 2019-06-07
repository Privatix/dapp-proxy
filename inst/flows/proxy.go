package flows

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/privatix/dappctrl/util"

	installerutil "github.com/privatix/dapp-installer/util"
)

type prodDirPath struct {
	DataDir                   string
	V2RayAgentConfig          string
	V2RayClientConfig         string
	V2RayExec                 string
	PluginExec                string
	PluginAgentConf           string
	PluginClientConf          string
	PluginAgentConfTpl        string
	PluginClientConfTpl       string
	OSXFirewallScript         string
	OSXFirewallRuleFile       string
	OSXConfigureProxyScript   string
	WINFirewallScript         string
	WinConfigureProxyScript   string
	LinuxFirewallScript       string
	LinuxConfigureProxyScript string
	OSXSyncTimeScript         string
}

// ProxyInstallation is proxy product installation details.
type ProxyInstallation struct {
	IsAgent          bool
	ProdDir          string
	ProdDirToUpdate  string
	V2RayDaemonName  string
	V2RayDaemonDesc  string
	PluginDaemonName string
	PluginDaemonDesc string

	Path prodDirPath
}

// NewProxyInstallation returns an instance with default values set.
func NewProxyInstallation() *ProxyInstallation {
	execPostfix := ""
	if runtime.GOOS == "windows" {
		execPostfix = ".exe"
	}
	return &ProxyInstallation{
		Path: prodDirPath{
			DataDir:                   "data",
			V2RayAgentConfig:          "config/agent.v2ray.config.json",
			V2RayClientConfig:         "config/client.v2ray.config.json",
			V2RayExec:                 "bin/v2ray/v2ray" + execPostfix,
			PluginExec:                "bin/dappproxy" + execPostfix,
			PluginAgentConf:           "config/adapter.agent.config.json",
			PluginClientConf:          "config/adapter.client.config.json",
			PluginAgentConfTpl:        "template/adapter.agent.config.json",
			PluginClientConfTpl:       "template/adapter.client.config.json",
			OSXFirewallScript:         "bin/scripts/mac/pf-rule.sh",
			OSXFirewallRuleFile:       "data/dapp-proxy.firewall.rule",
			OSXConfigureProxyScript:   "bin/scripts/mac/configuresocksfirewallproxy.sh",
			OSXSyncTimeScript:         "bin/mac/sync-time.sh",
			WINFirewallScript:         "bin/scripts/win/set-firewall-rule.ps1",
			WinConfigureProxyScript:   "bin/scripts/win/update-proxysettings.ps1",
			LinuxFirewallScript:       "bin/scripts/linux/configure-firewall.bash",
			LinuxConfigureProxyScript: "bin/scripts/linux/configure-socks-proxy.bash",
		},
	}
}

// init reads installation state or inits new.
func (p *ProxyInstallation) init(proddir, role string) error {
	err := p.setProdDir(proddir)
	if err != nil {
		return err
	}

	// Set role.
	p.IsAgent = role == roleAgent

	// Set daemon names and descriptions.
	h := installerutil.Hash(p.ProdDir)
	p.V2RayDaemonName = daemonName("v2ray", h)
	p.V2RayDaemonDesc = daemonDescription(p.role(), "v2ray", h)
	p.PluginDaemonName = daemonName("dappproxy", h)
	p.PluginDaemonDesc = daemonDescription(p.role(), "dappproxy", h)

	return nil
}

func (p *ProxyInstallation) setProdDir(dir string) error {
	if strings.HasPrefix(dir, ".") {
		dir = filepath.Join(filepath.Dir(os.Args[0]), dir)
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p.ProdDir = filepath.ToSlash(path)

	return nil
}

func (p *ProxyInstallation) saveAsFile() error {
	f, err := os.Create(p.installationFile())
	if err != nil {
		return fmt.Errorf("could not create installation file: %v", err)
	}

	return json.NewEncoder(f).Encode(p)
}

func (p *ProxyInstallation) readInstallationDetails() error {
	return util.ReadJSONFile(p.installationFile(), p)
}

func (p *ProxyInstallation) installationFile() string {
	return filepath.Join(p.ProdDir, "config/.env.config.json")
}

func (p *ProxyInstallation) role() string {
	if p.IsAgent {
		return "agent"
	}
	return "client"
}

func daemonName(ex, h string) string {
	return fmt.Sprintf("io.privatix.%s_%s", ex, h)
}

func daemonDescription(role, name, h string) string {
	return fmt.Sprintf("Privatix %s %s %s", role, name, h)
}

func (p *ProxyInstallation) prodPathJoin(f string) string {
	return filepath.Join(p.ProdDir, f)
}

func (p *ProxyInstallation) prodPathToUpdateJoin(f string) string {
	return filepath.Join(p.ProdDirToUpdate, f)
}

func (p *ProxyInstallation) v2rayExecPath() string {
	return p.prodPathJoin(p.Path.V2RayExec)
}

func (p *ProxyInstallation) v2rayConfPath() string {
	if p.IsAgent {
		return p.prodPathJoin(p.Path.V2RayAgentConfig)
	}
	return p.prodPathJoin(p.Path.V2RayClientConfig)
}

func (p *ProxyInstallation) pluginExecPath() string {
	return p.prodPathJoin(p.Path.PluginExec)
}

func (p *ProxyInstallation) pluginConfPath() string {
	if p.IsAgent {
		return p.prodPathJoin(p.Path.PluginAgentConf)
	}
	return p.prodPathJoin(p.Path.PluginClientConf)
}

func (p *ProxyInstallation) pluginClientConfTplPath() string {
	return p.prodPathJoin(p.Path.PluginClientConfTpl)
}

func (p *ProxyInstallation) pluginClientConfTplPathToUpdate() string {
	return p.prodPathToUpdateJoin(p.Path.PluginClientConfTpl)
}

func (p *ProxyInstallation) pluginAgentConfigTplPath() string {
	return p.prodPathJoin(p.Path.PluginAgentConfTpl)
}

func (p *ProxyInstallation) pluginAgentConfigTplPathToUpdate() string {
	return p.prodPathToUpdateJoin(p.Path.PluginAgentConfTpl)
}

func (p *ProxyInstallation) pluginAgentConfigPath() string {
	return p.prodPathJoin(p.Path.PluginAgentConf)
}

func (p *ProxyInstallation) pluginAgentConfigPathToUpdate() string {
	return p.prodPathToUpdateJoin(p.Path.PluginAgentConf)
}

func (p *ProxyInstallation) pluginClientConfigPath() string {
	return p.prodPathJoin(p.Path.PluginClientConf)
}

func (p *ProxyInstallation) pluginClientConfigPathToUpdate() string {
	return p.prodPathToUpdateJoin(p.Path.PluginClientConf)
}

func (p *ProxyInstallation) configureProxyScript() string {
	if goos := runtime.GOOS; goos == "darwin" {
		return p.prodPathJoin(p.Path.OSXConfigureProxyScript)
	} else if goos == "windows" {
		return p.prodPathJoin(p.Path.WinConfigureProxyScript)
	} else if goos == "linux" {
		return p.prodPathJoin(p.Path.LinuxConfigureProxyScript)
	}
	return ""
}

func (p *ProxyInstallation) logsDirPath() string {
	return p.prodPathJoin("log")
}

func (p *ProxyInstallation) dataDirPath() string {
	return p.prodPathJoin("data")
}

func (p *ProxyInstallation) osxFilrewallScript() string {
	return p.prodPathJoin(p.Path.OSXFirewallScript)
}

func (p *ProxyInstallation) osxFirewallRuleFile() string {
	return p.prodPathJoin(p.Path.OSXFirewallRuleFile)
}

func (p *ProxyInstallation) winFirewallScript() string {
	return p.prodPathJoin(p.Path.WINFirewallScript)
}

func (p *ProxyInstallation) syncTimeScriptPath() string {
	return p.prodPathJoin(p.Path.OSXSyncTimeScript)
}

func (p *ProxyInstallation) linuxFirewallScript() string {
	return p.prodPathJoin(p.Path.LinuxFirewallScript)
}
