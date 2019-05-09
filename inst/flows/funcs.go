package flows

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/takama/daemon"

	"github.com/privatix/dapp-installer/util"
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
	// TODO: Check there is not active installation.
	return nil
}

func readJSON(f string) (map[string]interface{}, error) {
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(dat, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func setLogPath(p *ProxyInstallation, m map[string]interface{}) {
	newflog := m["FileLog"].(map[string]interface{})

	newflog["Filename"] = filepath.Join(p.logsDirPath(), "dappproxy-%Y-%m-%d.log")

	m["FileLog"] = newflog
}

func saveToFile(m map[string]interface{}, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(m)
}

func preparePluginConfigs(p *ProxyInstallation) error {
	if p.IsAgent {
		m, err := readJSON(p.pluginAgentConfigTplPath())
		if err != nil {
			return err
		}

		setLogPath(p, m)

		return saveToFile(m, p.pluginAgentConfigPath())
	}

	m, err := readJSON(p.pluginClientConfTplPath())
	if err != nil {
		return err
	}

	setLogPath(p, m)

	m["ConfigureProxyScript"] = p.configureProxyScript()

	return saveToFile(m, p.pluginClientConfigPath())
}

func prepareUpdateFromPluginConfigs(p *ProxyInstallation) error {
	if p.IsAgent {
		m, err := readJSON(p.pluginAgentConfigTplPathToUpdateFrom())
		if err != nil {
			return err
		}

		setLogPath(p, m)

		return saveToFile(m, p.pluginAgentConfigPathToUpdateFrom())
	}

	m, err := readJSON(p.pluginClientConfTplPathToUpdateFrom())
	if err != nil {
		return err
	}

	setLogPath(p, m)

	m["ConfigureProxyScript"] = p.configureProxyScript()

	return saveToFile(m, p.pluginClientConfigPathToUpdateFrom())
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
	return parseProdDirOrHelpFlags(p, helpRemove)
}

func parseUpdateFlags(p *ProxyInstallation) error {
	return parseProdDirOrHelpFlags(p, helpUpdate)
}

func locateProductTempDir(p *ProxyInstallation) error {
	productTempPath := os.Getenv("PRIVATIX_TEMP_PRODUCT")
	_, uuidProd := filepath.Split(p.ProdDir)

	if productTempPath == "" {
		return fmt.Errorf("PRIVATIX_TEMP_PRODUCT is empty")
	}

	found := ""

	err := filepath.Walk(productTempPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return err
		}
		_, dir := filepath.Split(path)

		if err == nil && strings.EqualFold(dir, uuidProd) {
			found = path
		}

		return err
	})
	if err != nil {
		return err
	}

	if found == "" {
		return fmt.Errorf("could not find product dir '%s' at PRIVATIX_TEMP_PRODUCT", uuidProd)
	}

	p.ProdDirToUpdateFrom = found

	return nil
}

func copyDataDirFiles(p *ProxyInstallation) error {
	return util.CopyDir(p.prodPathJoin(p.Path.DataDir), p.prodPathToUpdateFromJoin(p.Path.DataDir))
}

func copyAndMergeConfigs(p *ProxyInstallation) error {
	// TODO: implement.
	return nil
}

func parseV2RayRunFlags(p *ProxyInstallation) error {
	return parseCommonFlags(p, MethodRunV2Ray)
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
