package adapter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	ipify "github.com/rdegges/go-ipify"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
)

func newProductConfig(conf V2RayConfig) map[string]string {
	m := make(map[string]string)
	m[productAlterID] = fmt.Sprint(conf.AlterID)
	addr, err := ipify.GetIp()
	must("couldn't get my IP address", err)
	m[sess.ProductExternalIP] = addr
	m[productAddress] = addr
	m[productPort] = fmt.Sprint(conf.InboundPort)
	return m
}

func pushConfiguration(conf V2RayConfig) {
	params := newProductConfig(conf)
	err := adapterSessClient.SetProductConfig(params)
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
func AsAgent(conf *Config, workdir string) {
	beforeStart := func() {
		if configNotPushed(workdir) {
			pushConfiguration(conf.V2Ray)
			markConfigAsPushed(workdir)
		}
	}
	onConnCreate := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("configuring proxy to accept connection")
		err := adapterUsersClient.AddUser(context.Background(), *endpoint.Username)
		must("", err)
		adapterMon.Start(*endpoint.Username, change.Channel)
	}
	onConnStart := func(_ *data.Endpoint, _ *sess.ConnChangeResult) {}
	onConnStop := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("configuring proxy to close connection")
		err := adapterUsersClient.RemoveUser(context.Background(), *endpoint.Username)
		must("", err)
		adapterMon.Stop(*endpoint.Username, change.Channel)
	}

	runAdapter(conf, beforeStart, onConnCreate, onConnStart, onConnStop)
}
