package flows

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/privatix/dapp-proxy/plugin/adapter"
)

func TestInstallFuncs(t *testing.T) {
	t.Run("ValidateInstallEnvironment", func(t *testing.T) {
		dir := createTempDirOrFail(t)
		defer os.RemoveAll(dir)

		p := NewProxyInstallation()
		p.init(dir, "agent")

		t.Run("OK", func(t *testing.T) {
			// No ".env.config.json" file is present means ok.
			if err := validateInstallEnvironment(p); err != nil {
				t.Fatalf("validateInstallEnvironment() returned error: %v", err)
			}
		})

		t.Run("Error", func(t *testing.T) {
			createSubdirOrFail(t, dir, "config")
			createFileOrFail(t, dir, "config/.env.config.json")
			// ".env.config.json" file is created, validation must fail.
			if err := validateInstallEnvironment(p); err == nil {
				t.Fatalf("validateInstallEnvironment() did not validate")
			}
		})
	})

	t.Run("InstallStepsInOrder", func(t *testing.T) {
		// Test installation funcs in flow order.
		// Some funcs are excluded in tests due to mocking complexity and/or logic simplicity.
		// Testing funcs:
		//    (ignored) parseInstallFlags
		// 1. preparePluginConfigs
		//    (ignored) createV2RayDaemon
		//    (ignored) createPluginDaemon
		// 2. configureOSFirewall
		proddir := createTempDirOrFail(t)
		defer os.RemoveAll(proddir)

		// Proxy configuration object is used in all funcs tests here.
		p := NewProxyInstallation()
		// Client installation is the same as agent, but agent has a bit more
		// things to do. So testing agent tests client too.
		p.init(proddir, "agent")

		// Copy source dir.
		copyDirOrFail(t, os.Getenv("SOURCE_DIR"), proddir)

		// Test all funcs in order.

		testPreparePluginConfigs(t, p)
		testConfigureOSFirewall(t, p)
	})
}

func testPreparePluginConfigs(t *testing.T, p *ProxyInstallation) {
	// Must copy plugin config files and adjust pathes in it.

	if err := preparePluginConfigs(p); err != nil {
		t.Fatalf("preparePluginConfigs(p) returned error: %v", err)
	}
	// Check pathes are updated.

	// Client config.
	configContent := ReadFileOrFail(t, p.pluginClientConfigPath())
	if !strings.Contains(configContent, filepath.Join(p.ProdDir, "log/dappproxy-%Y-%m-%d.log")) {
		t.Fatal("client log path is not updated")
	}
	if !strings.Contains(configContent, filepath.Join(p.ProdDir, "data")) {
		t.Fatal("client config `ChannelDir` not updated")
	}
	if !strings.Contains(configContent, p.configureProxyScript()) {
		t.Fatal("configure proxy script path not updated")
	}

	// Agent config.
	configContent = ReadFileOrFail(t, p.pluginAgentConfigPath())
	if !strings.Contains(configContent, filepath.Join(p.ProdDir, "log/dappproxy-%Y-%m-%d.log")) {
		t.Fatal("agent log path is not updated")
	}
	if !strings.Contains(configContent, filepath.Join(p.ProdDir, "data")) {
		t.Fatal("agent config `ChannelDir` not updated")
	}
}

func testConfigureOSFirewall(t *testing.T, p *ProxyInstallation) {
	// configureOSFirewall.
	// Must configure os firewall by building rule file from template,
	// putting it in data directory and executing it.

	// Fake the script file.
	fakescript := filepath.Join(p.ProdDir, "data/fake-script.sh")
	outfile := filepath.Join(p.ProdDir, "script-output")
	// To test that script was properly built and executed the script output
	// file tested for containing correct arguments.
	writeFileOrFail(t, fakescript, []byte(fmt.Sprintf("#!/bin/sh\necho $@ >  %s", outfile)))

	p.Path.OSXFirewallScript = "data/fake-script.sh"

	// Set test v2ray port.
	var adapterConfig adapter.Config
	readJSONORFail(t, p.pluginAgentConfigPath(), &adapterConfig)
	adapterConfig.V2Ray.InboundPort = 1234
	writeJSONOrFail(t, p.pluginAgentConfigPath(), &adapterConfig)

	if err := configureOSFirewall(p); err != nil {
		t.Fatalf("configureOSFirewall returned error: %v", err)
	}
	// Check firewall script executed by examining output file.
	if content, err := ioutil.ReadFile(outfile); err != nil {
		t.Fatalf("failed to read firewall rule file: %v", err)
	} else if exp := fmt.Sprint("on ", adapterConfig.V2Ray.InboundPort, " ", p.osxFirewallRuleFile()); !strings.Contains(string(content), exp) {
		t.Fatalf("`%s` not found in firewall rule file: %s", exp, content)
	}
}

func createTempDirOrFail(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "dapp-product")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func createSubdirOrFail(t *testing.T, dir, subdir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, subdir), os.ModePerm); err != nil {
		t.Fatal(err)
	}
}

func createFileOrFail(t *testing.T, dir, file string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, file))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func copyFileOrFail(t *testing.T, src, dst string) {
	t.Helper()

	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		t.Fatal(err)
	}
}

func copyDirOrFail(t *testing.T, src, dst string) {
	si, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		t.Fatal(err)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			copyDirOrFail(t, srcPath, dstPath)
		} else {
			copyFileOrFail(t, srcPath, dstPath)
		}
	}

	return
}

func ReadFileOrFail(t *testing.T, file string) string {
	t.Helper()

	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func writeFileOrFail(t *testing.T, file string, content []byte) {
	t.Helper()

	if err := ioutil.WriteFile(file, content, os.ModePerm); err != nil {
		t.Fatal(err)
	}
}

func readJSONORFail(t *testing.T, file string, out interface{}) {
	t.Helper()

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewDecoder(f).Decode(&out); err != nil {
		t.Fatal(err)
	}
}

func writeJSONOrFail(t *testing.T, file string, v interface{}) {
	t.Helper()

	f, err := os.OpenFile(file, os.O_WRONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(f).Encode(v); err != nil {
		t.Fatal(err)
	}
}
