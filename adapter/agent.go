package adapter

import (
	"context"

	"github.com/privatix/dappctrl/sess"
)

// AsAgent runs adapter in agent mode.
func AsAgent() {
	conf := &agentConfig{}
	readConfigFile(conf)

	client := newV2RayClient(conf.V2Ray.API, conf.V2Ray.InboundTag, conf.V2Ray.AlterID)

	sesscl := dialSess(conf.Sess)

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
