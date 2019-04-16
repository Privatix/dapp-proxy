package flows

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/takama/daemon"
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

func setLogPathAndCopy(p *ProxyInstallation, tpl, dest string) error {
	dat, err := ioutil.ReadFile(tpl)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(dat, &m)
	if err != nil {
		return err
	}

	newflog := m["FileLog"].(map[string]interface{})

	newflog["Filename"] = filepath.Join(p.logsDirPath(), "dappproxy-%Y-%m-%d.log")

	m["FileLog"] = newflog

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(m)
}

func preparePluginConfigs(p *ProxyInstallation) error {
	err := setLogPathAndCopy(p, p.pluginAgentConfigTplPath(), p.pluginAgentConfigPath())
	if err != nil {
		return err
	}
	return setLogPathAndCopy(p, p.pluginClientConfTplPath(), p.pluginClientConfigPath())
}

func createDaemons(p *ProxyInstallation) error {
	err := createV2RayDaemon(p)
	if err == nil {
		err = createAdapterDaemon(p)
	}
	return err
}

func removeDaemons(p *ProxyInstallation) error {
	err := removeAdapterDaemon(p)
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

func createAdapterDaemon(p *ProxyInstallation) error {
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

func removeAdapterDaemon(p *ProxyInstallation) error {
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

func startDaemons(p *ProxyInstallation) error {
	err := startDaemon(p.V2RayDaemonName)
	if err == nil {
		err = startDaemon(p.PluginDaemonName)
	}
	return err
}

func stopDaemonsSilent(p *ProxyInstallation) error {
	for _, name := range []string{p.V2RayDaemonName, p.PluginDaemonName} {
		service, err := daemon.New(name, "")
		if err != nil {
			return fmt.Errorf("failed to get '%s' daemon: %v", name, err)
		}

		service.Stop()
	}

	return nil
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

func parseRemoveFlags(p *ProxyInstallation) error {
	h := flag.Bool("help", false, "Display installer help")
	proddir := flag.String("proddir", "", "Product install directory")

	flag.CommandLine.Parse(os.Args[2:])

	if *h || *proddir == "" {
		fmt.Println(helpRemove)
		os.Exit(0)
	}

	p.setProdDir(*proddir)

	return nil
}

func checkInstallation(p *ProxyInstallation) error {
	// TODO: implement.
	return nil
}

func parseUpdateFlags(p *ProxyInstallation) error {
	// TODO: implement.
	return nil
}

func copyDataDir(p *ProxyInstallation) error {
	// TODO: implement.
	return nil
}

func parseStartFlags(p *ProxyInstallation) error {
	// TODO: implement.
	return nil
}

func parseStopFlags(p *ProxyInstallation) error {
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
