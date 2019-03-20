package adapter

import (
	"github.com/privatix/dappctrl/sess"
)

// AsClient runs adapter in clients mode.
func AsClient() {
	conf := new(clientConfig)
	readConfigFile(conf)

	sesscl := newProductSessClient(conf.Sess)

	changesChan := connChangeSubscribe(sesscl)

	client := newV2RayStatsClient(conf.V2Ray.API, conf.V2Ray.InboundTag)

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
			// 1. Run/configure v2ray.
			// 2. Start monitoring.
			mon.Start(username)
			// 3. Start reading v2ray logs to detect connection drops.
			// How to recognize logs particularly for this connection?
		case sess.ConnStop:
			// 1. Stop or restore v2ray configuration.
			// 2. Stop reading v2ray logs for this connection.
			mon.Stop(username)
		}
	}
}
