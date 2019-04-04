package main

import (
	"flag"
	"path/filepath"

	"github.com/privatix/dapp-proxy/plugin/adapter"
	"github.com/privatix/dappctrl/util"
)

func readConf(conf interface{}, confFile string) {
	err := util.ReadJSONFile(confFile, &conf)
	if err != nil {
		panic("failed to read configuration: " + err.Error())
	}
}

func main() {
	fconfig := flag.String("config", "config.json", "Configuration file")
	flag.Parse()

	conf := new(adapter.Config)
	readConf(conf, *fconfig)

	// If configuration is valid for agent then start agent.
	// Otherwise start client.
	if adapter.ValidAgentConf(conf) {
		workdir := filepath.Dir(*fconfig)
		adapter.AsAgent(conf, workdir)
		return
	}

	adapter.AsClient(conf)
}
