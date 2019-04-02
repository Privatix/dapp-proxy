package mode

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	ipify "github.com/rdegges/go-ipify"

	"github.com/privatix/dappctrl/sess"
)

func newProductConfig(conf V2RayAgentConfig) map[string]string {
	m := make(map[string]string)
	m[productAlterID] = fmt.Sprint(conf.AlterID)
	addr, err := ipify.GetIp()
	must("couldn't get my IP address", err)
	m[sess.ProductExternalIP] = addr
	m[productAddress] = addr
	m[productPort] = fmt.Sprint(conf.InboundPort)
	return m
}

func pushConfiguration(conf V2RayAgentConfig, sesscl *sess.Client) {
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
func AsAgent(conf *AgentConfig, workdir string) {

	sesscl := newProductSessClient(conf.Sess)

	if configNotPushed(workdir) {
		pushConfiguration(conf.V2Ray, sesscl)
		markConfigAsPushed(workdir)
	}

	conn := newV2RayAPIConn(conf.V2Ray.API)

	statsclient := newV2RayStatsClient(conn, conf.V2Ray.InboundTag)

	usersclient := newV2RayUsersClient(conn, conf.V2Ray.InboundTag,
		conf.V2Ray.AlterID)

	changesChan := connChangeSubscribe(sesscl)

	logger, closer := createLogger(conf.FileLog)
	defer closer.Close()

	mon := newMonitor(statsclient, conf.Monitor, logger)

	go handleReports(mon, sesscl, logger)

	for change := range changesChan {
		logger := logger.Add("connectionChange", *change)

		endpoint, err := sesscl.GetEndpoint(change.Channel)
		if err != nil {
			logger.Fatal(err.Error())
		}

		logger = logger.Add("endpoint", *endpoint)

		if endpoint.Username == nil {
			logger.Fatal("username of connection change is empty")
		}

		username := *endpoint.Username

		switch change.Status {
		case sess.ConnCreate:
			err = usersclient.AddUser(context.Background(), username)
			must("", err)
			mon.Start(username, change.Channel)
		case sess.ConnStop:
			err = usersclient.RemoveUser(context.Background(), username)
			must("", err)
			mon.Stop(username, change.Channel)
		}
	}
}
