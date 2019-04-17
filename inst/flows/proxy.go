package flows

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dappctrl/util"
)

const (
	installationFilename = "installation.json"
)

type prodDirPath struct {
	V2RayAgentConfig    string
	V2RayClientConfig   string
	V2RayExec           string
	PluginExec          string
	PluginAgentConf     string
	PluginClientConf    string
	PluginAgentConfTpl  string
	PluginClientConfTpl string
}

// ProxyInstallation is proxy product installation details.
type ProxyInstallation struct {
	IsAgent          bool
	ProdDir          string
	V2RayDaemonName  string
	V2RayDaemonDesc  string
	PluginDaemonName string
	PluginDaemonDesc string

	Path prodDirPath
}

// NewProxyInstallation returns an instance with default values set.
func NewProxyInstallation() *ProxyInstallation {
	return &ProxyInstallation{
		Path: prodDirPath{
			V2RayAgentConfig:    "config/agent.v2ray.config.json",
			V2RayClientConfig:   "config/client.v2ray.config.json",
			V2RayExec:           "bin/v2ray/v2ray",
			PluginExec:          "bin/dappproxy",
			PluginAgentConf:     "config/adapter.agent.config.json",
			PluginClientConf:    "config/adapter.client.config.json",
			PluginAgentConfTpl:  "template/adapter.agent.config.json",
			PluginClientConfTpl: "template/adapter.client.config.json",
		},
	}
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.ToLower(s)))
	return hex.EncodeToString(h.Sum(nil))
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
	h := hash(p.ProdDir)
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
	f, err := os.Create(filepath.Join(p.ProdDir, "data", installationFilename))
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(p)
}

func (p *ProxyInstallation) readInstallationDetails() error {
	return util.ReadJSONFile(filepath.Join(p.ProdDir, "data", installationFilename), p)
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

func (p *ProxyInstallation) pathJoin(f string) string {
	return filepath.Join(p.ProdDir, f)
}

func (p *ProxyInstallation) v2rayExecPath() string {
	return p.pathJoin(p.Path.V2RayExec)
}

func (p *ProxyInstallation) v2rayConfPath() string {
	if p.IsAgent {
		return p.pathJoin(p.Path.V2RayAgentConfig)
	}
	return p.pathJoin(p.Path.V2RayClientConfig)
}

func (p *ProxyInstallation) pluginExecPath() string {
	return p.pathJoin(p.Path.PluginExec)
}

func (p *ProxyInstallation) pluginConfPath() string {
	if p.IsAgent {
		return p.pathJoin(p.Path.PluginAgentConf)
	}
	return p.pathJoin(p.Path.PluginClientConf)
}

func (p *ProxyInstallation) pluginClientConfTplPath() string {
	return p.pathJoin(p.Path.PluginClientConfTpl)
}

func (p *ProxyInstallation) pluginAgentConfigTplPath() string {
	return p.pathJoin(p.Path.PluginAgentConfTpl)
}

func (p *ProxyInstallation) pluginAgentConfigPath() string {
	return p.pathJoin(p.Path.PluginAgentConf)
}

func (p *ProxyInstallation) pluginClientConfigPath() string {
	return p.pathJoin(p.Path.PluginClientConf)
}

func (p *ProxyInstallation) logsDirPath() string {
	return p.pathJoin("log")
}
