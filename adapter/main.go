package main

import (
	"flag"
	"path/filepath"

	"github.com/privatix/dapp-proxy/adapter/flow"
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

	agentConf := new(flow.AgentConfig)
	readConf(agentConf, *fconfig)
	if flow.ValidAgentConf(agentConf) {
		workdir := filepath.Dir(*fconfig)
		flow.AsAgent(agentConf, workdir)
		return
	}

	// If configuration is not valid for agent, it's must be running for client.
	conf := new(flow.ClientConfig)
	readConf(conf, *fconfig)
	flow.AsClient(conf)
}
