package adapter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	ipify "github.com/rdegges/go-ipify"

	"github.com/privatix/dappctrl/sess"
)

func newProductConfig(conf v2rayAgentConfig) map[string]string {
	m := make(map[string]string)
	m[productAlterID] = fmt.Sprint(conf.AlterID)
	addr, err := ipify.GetIp()
	must("couldn't get my IP address", err)
	m[sess.ProductExternalIP] = addr
	m[productAddress] = addr
	m[productPort] = fmt.Sprint(conf.InboundPort)
	return m
}

func pushConfiguration(conf v2rayAgentConfig, sesscl *sess.Client) {
	params := newProductConfig(conf)
	err := sesscl.SetProductConfig(params)
	must("could not push product configiration", err)
}

// Config push file related constants.
const (
	pushedFile = "configPushed"
	filePerm   = 0644
)

func configNotPushed(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, pushedFile))
	return err != nil
}

func markConfigAsPushed(dir string) {
	file := filepath.Join(dir, pushedFile)
	err := ioutil.WriteFile(file, nil, filePerm)
	must("could not mark product config as pushed", err)
}

// AsAgent runs adapter in agent mode.
func AsAgent() {
	conf := &agentConfig{}

	confFile := readConfigFile(conf)

	sesscl := newProductSessClient(conf.Sess)

	dir := filepath.Dir(confFile)
	if configNotPushed(dir) {
		pushConfiguration(conf.V2Ray, sesscl)
		markConfigAsPushed(dir)
	}

	statsclient := newV2RayStatsClient(conf.V2Ray.API, conf.V2Ray.InboundTag)

	usersclient := newV2RayUsersClient(conf.V2Ray.API, conf.V2Ray.InboundTag,
		conf.V2Ray.AlterID)

	changesChan := connChangeSubscribe(sesscl)

	mon := newMonitor(statsclient, conf.Monitor)

	go handleReports(mon, sesscl)

	for change := range changesChan {
		endpoint, err := sesscl.GetEndpoint(change.Channel)
		must("", err)

		if endpoint.Username == nil {
			// TODO: log error or fatal.
		}

		username := *endpoint.Username

		switch change.Status {
		case sess.ConnStart:
			err = usersclient.AddUser(context.Background(), username)
			must("", err)
			mon.Start(username)
		case sess.ConnStop:
			err = usersclient.RemoveUser(context.Background(), username)
			must("", err)
			mon.Stop(username)
		}
	}
}
