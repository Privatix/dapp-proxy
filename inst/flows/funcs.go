package flows

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/takama/daemon"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-proxy/plugin/adapter"
)

const (
	roleAgent  = "agent"
	roleClient = "client"
)

func parseCommonFlags(p *ProxyInstallation, method string) error {
	h := flag.Bool("help", false, "Display installer help")
	role := flag.String("role", "", "Product role")
	proddir := flag.String("proddir", "", "Product install directory")

	flag.CommandLine.Parse(os.Args[2:])

	if *role != "" && *role != roleClient && *role != roleAgent {
		return fmt.Errorf("--role can be either 'client' or 'agent'")
	}

	if *h || *role == "" || *proddir == "" {
		fmt.Printf(commonHelpFormat, method)
		os.Exit(0)
	}

	return p.init(*proddir, *role)
}

func parseInstallFlags(p *ProxyInstallation) error {
	return parseCommonFlags(p, MethodInstall)
}

func validateInstallEnvironment(p *ProxyInstallation) error {
	_, err := os.Stat(p.installationFile())
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("already installed")
}

func setLogPath(p *ProxyInstallation, config *adapter.Config) {
	config.FileLog.Filename = filepath.Join(p.logsDirPath(), "dappproxy-%Y-%m-%d.log")
}

func setChannelDir(p *ProxyInstallation, v *adapter.Config) {
	v.ChannelDir = p.dataDirPath()
}

func saveJSON(v interface{}, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(v)
}

func preparePluginConfigs(p *ProxyInstallation) error {
	config := new(adapter.Config)
	if err := readJSON(p.pluginAgentConfigTplPath(), config); err != nil {
		return err
	}

	setLogPath(p, config)

	setChannelDir(p, config)

	if err := saveJSON(config, p.pluginAgentConfigPath()); err != nil {
		return err
	}

	if err := readJSON(p.pluginClientConfTplPath(), config); err != nil {
		return err
	}

	setLogPath(p, config)

	setChannelDir(p, config)

	config.ConfigureProxyScript = p.configureProxyScript()

	return saveJSON(config, p.pluginClientConfigPath())
}

func prepareUpdatePluginConfigs(p *ProxyInstallation) error {
	config := new(adapter.Config)
	if p.IsAgent {
		if err := readJSON(p.pluginAgentConfigTplPathToUpdate(), config); err != nil {
			return err
		}

		setLogPath(p, config)

		setChannelDir(p, config)

		return saveJSON(config, p.pluginAgentConfigPathToUpdate())
	}

	if err := readJSON(p.pluginClientConfTplPathToUpdate(), config); err != nil {
		return err
	}

	setLogPath(p, config)

	setChannelDir(p, config)

	config.ConfigureProxyScript = p.configureProxyScript()

	return saveJSON(config, p.pluginClientConfigPathToUpdate())
}

func removeDaemons(p *ProxyInstallation) error {
	err := removePluginDaemon(p)
	if err == nil {
		err = removeV2RayDaemon(p)
	}
	return err
}

func createV2RayDaemon(p *ProxyInstallation) error {
	service, err := daemon.New(p.V2RayDaemonName, p.V2RayDaemonDesc)
	if err != nil {
		return fmt.Errorf("failed to create v2ray daemon: %v", err)
	}

	_, err = service.Install("run-v2ray", "--proddir", p.ProdDir, "--role", p.role())
	if err != nil {
		return fmt.Errorf("failed to install %s daemon: %v", p.V2RayDaemonName, err)
	}

	return nil
}

func removeV2RayDaemon(p *ProxyInstallation) error {
	service, err := daemon.New(p.V2RayDaemonName, "")
	if err != nil {
		return fmt.Errorf("failed to get v2ray daemon: %v", err)
	}

	_, err = service.Remove()
	if err != nil {
		return fmt.Errorf("failed to remove v2ray daemon: %v", err)
	}

	return nil
}

func createPluginDaemon(p *ProxyInstallation) error {
	service, err := daemon.New(p.PluginDaemonName, p.PluginDaemonDesc, p.V2RayDaemonName)
	if err != nil {
		return fmt.Errorf("failed to create adapter daemon: %v", err)
	}

	_, err = service.Install("run-plugin", "--proddir", p.ProdDir, "--role", p.role())
	if err != nil {
		return fmt.Errorf("failed to install %s: %v", p.PluginDaemonName, err)
	}

	return nil
}

func removePluginDaemon(p *ProxyInstallation) error {
	service, err := daemon.New(p.PluginDaemonName, "")
	if err != nil {
		return fmt.Errorf("failed to get adapter daemon: %v", err)
	}

	_, err = service.Remove()
	if err != nil {
		return fmt.Errorf("failed to remove adapter '%s': %v", p.PluginDaemonName, err)
	}

	return nil
}

func startDaemonsSilent(p *ProxyInstallation) error {
	for _, name := range []string{p.V2RayDaemonName, p.PluginDaemonName} {
		service, err := daemon.New(name, "")
		if err != nil {
			return fmt.Errorf("failed to get '%s' daemon: %v", name, err)
		}

		service.Start()
	}

	return nil
}

func startDaemon(name string) error {
	service, err := daemon.New(name, "")
	if err != nil {
		return fmt.Errorf("failed to get '%s' daemon: %v", name, err)
	}

	_, err = service.Start()
	if err != nil {
		return fmt.Errorf("failed to start '%s': %v", name, err)
	}

	return nil
}

func startV2rayDaemon(p *ProxyInstallation) error {
	return startDaemon(p.V2RayDaemonName)
}

func startPluginDaemon(p *ProxyInstallation) error {
	return startDaemon(p.PluginDaemonName)
}

func stopV2rayDaemon(p *ProxyInstallation) error {
	return stopDaemon(p.V2RayDaemonName)
}

func stopPluginDaemon(p *ProxyInstallation) error {
	return stopDaemon(p.PluginDaemonName)
}

func stopDaemon(name string) error {
	service, err := daemon.New(name, "")
	if err != nil {
		return fmt.Errorf("failed to get '%s' daemon: %v", name, err)
	}

	_, err = service.Stop()
	return err
}

func stopDaemons(p *ProxyInstallation) error {
	for _, name := range []string{p.V2RayDaemonName, p.PluginDaemonName} {
		service, err := daemon.New(name, "")
		if err != nil {
			return fmt.Errorf("failed to get '%s' daemon: %v", name, err)
		}

		_, err = service.Stop()
		if err != nil {
			return fmt.Errorf("failed to stop '%s' daemon: %v", name, err)
		}
	}

	return nil
}

func saveInstallationDetails(p *ProxyInstallation) error {
	return p.saveAsFile()
}

func readInstallationDetails(p *ProxyInstallation) error {
	return p.readInstallationDetails()
}

func parseProdDirOrHelpFlags(p *ProxyInstallation, helpMsg string) error {
	h := flag.Bool("help", false, "Display installer help")
	proddir := flag.String("proddir", "", "Product install directory")

	flag.CommandLine.Parse(os.Args[2:])

	if *h || *proddir == "" {
		fmt.Println(helpMsg)
		os.Exit(0)
	}

	p.setProdDir(*proddir)

	return nil
}

func parseRemoveFlags(p *ProxyInstallation) error {
	return parseProdDirOrHelpFlags(p, fmt.Sprintf(commonHelpNoRole, MethodRemove))
}

func parseUpdateFlags(p *ProxyInstallation) error {
	return parseProdDirOrHelpFlags(p, fmt.Sprintf(commonHelpNoRole, MethodUpdate))
}

func locateProductDirToUpdate(p *ProxyInstallation) error {
	// HACK: It is known that dapp-installer creates a `role`_new dir
	// during update. Hardcoding this is burrow for bugs.
	prodNewLocation := strings.Replace(p.ProdDir, p.role()+"_new", p.role(), 1)

	_, uuidProd := filepath.Split(p.ProdDir)

	productTempPath := os.Getenv("PRIVATIX_TEMP_PRODUCT")

	if productTempPath != "" {
		// If dapp-installer provided PRIVATIX_TEMP_PRODUCT, search for
		// dir to update in there.
		err := filepath.Walk(productTempPath, func(name string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return err
			}
			_, dir := filepath.Split(name)

			if err == nil && strings.EqualFold(dir, uuidProd) {
				prodNewLocation = name
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	p.ProdDirToUpdate = prodNewLocation

	return nil
}

func copyDataDirFiles(p *ProxyInstallation) error {
	return util.CopyDir(p.prodPathToUpdateJoin(p.Path.DataDir), p.prodPathJoin(p.Path.DataDir))
}

func copyAndMergeConfigs(p *ProxyInstallation) error {
	if err := mergeConfigs(p.pluginAgentConfigPath(), p.pluginAgentConfigPathToUpdate()); err != nil {
		return err
	}
	return mergeConfigs(p.pluginClientConfigPath(), p.pluginClientConfigPathToUpdate())
}

func mergeConfigs(from, to string) error {
	var mapFrom, mapTo map[string]interface{}
	if err := readJSON(from, mapFrom); err != nil {
		return err
	}
	if err := readJSON(to, mapTo); err != nil {
		return err
	}

	return copyMissedKeys(mapFrom, mapTo)
}

func readJSON(file string, out interface{}) error {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return err
	}

	return json.NewDecoder(f).Decode(&out)
}

// copyMissedKeys copies missed keys that exist in `from` map to `to` map.
func copyMissedKeys(from map[string]interface{}, to map[string]interface{}) error {
	for k, v := range from {
		// If value is absent in `to` then copy as is.
		if _, ok := to[k]; !ok {
			to[k] = v
			continue
		}
		// If value is a map, then apply recursion.
		if vMap, ok := v.(map[string]interface{}); ok {
			if err := copyMissedKeys(vMap, to[k].(map[string]interface{})); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseV2RayRunFlags(p *ProxyInstallation) error {
	return parseCommonFlags(p, MethodRunV2Ray)
}

func parseStartStopFalgs(p *ProxyInstallation) error {
	return parseProdDirOrHelpFlags(p, fmt.Sprintf(commonHelpNoRole, MethodRemove))
}

func runV2Ray(p *ProxyInstallation) error {
	service, err := daemon.New(p.V2RayDaemonName, "")
	if err != nil {
		return fmt.Errorf("failed to get v2ray daemon: %v", err)
	}

	_, err = service.Run(&daemonExecute{
		execPath: p.v2rayExecPath(),
		confPath: p.v2rayConfPath(),
	})
	if err != nil {
		return fmt.Errorf("failed to run v2ray: %v", err)
	}
	return nil
}

func parsePluginRunFlags(p *ProxyInstallation) error {
	return parseCommonFlags(p, MethodRunPlugin)
}

func runPlugin(p *ProxyInstallation) error {
	service, err := daemon.New(p.PluginDaemonName, "")
	if err != nil {
		return fmt.Errorf("failed to get plugin daemon: %v", err)
	}

	_, err = service.Run(&daemonExecute{
		execPath: p.pluginExecPath(),
		confPath: p.pluginConfPath(),
	})
	if err != nil {
		return fmt.Errorf("failed to run plugin: %v", err)
	}
	return nil
}
