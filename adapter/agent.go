package adapter

import (
	"context"
	"fmt"

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

// AsAgent runs adapter in agent mode.
func AsAgent() {
	conf := &agentConfig{}
	readConfigFile(conf)

	sesscl := newProductSessClient(conf.Sess)

	pushConfiguration(conf.V2Ray, sesscl)

	client := newV2RayClient(conf.V2Ray.API, conf.V2Ray.InboundTag, conf.V2Ray.AlterID)

	changesChan := connChangeSubscribe(sesscl)

	mon := newMonitor(client, conf.Monitor)

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
			err = client.AddUser(context.Background(), username)
			must("", err)
			mon.Start(username)
		case sess.ConnStop:
			err = client.RemoveUser(context.Background(), username)
			must("", err)
			mon.Stop(username)
		}
	}
}
