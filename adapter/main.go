package main

import (
	"flag"
	"path/filepath"

	"github.com/privatix/dapp-proxy/adapter/mode"
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

	agentConf := new(mode.AgentConfig)
	readConf(agentConf, *fconfig)
	if mode.ValidAgentConf(agentConf) {
		workdir := filepath.Dir(*fconfig)
		mode.AsAgent(agentConf, workdir)
		return
	}

	// If configuration is not valid for agent, it's must be running for client.
	conf := new(mode.ClientConfig)
	readConf(conf, *fconfig)
	mode.AsClient(conf)
}
