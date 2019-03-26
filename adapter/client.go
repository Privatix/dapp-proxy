package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	v2rayclient "github.com/privatix/dapp-proxy/v2ray-client"
	"github.com/privatix/dappctrl/sess"
)

func newConfigureRequest(username string, prodConfRaw json.RawMessage) (*v2rayclient.VmessOutbound, error) {
	prodconf := make(map[string]string)

	err := json.Unmarshal(prodConfRaw, &prodconf)
	if err != nil {
		return nil, err
	}

	addr, ok := prodconf[productAddress]
	if !ok {
		return nil, fmt.Errorf("could not find %s", productAddress)
	}

	portRaw, ok := prodconf[productPort]
	if !ok {
		return nil, fmt.Errorf("could not find %s", productPort)
	}

	port, err := strconv.ParseUint(portRaw, 10, 8)
	if err != nil {
		return nil, err
	}

	alterIDRaw, ok := prodconf[productAlterID]
	if !ok {
		return nil, fmt.Errorf("could not find %s", productAlterID)
	}

	alterID, err := strconv.ParseUint(alterIDRaw, 10, 8)
	if err != nil {
		return nil, err
	}

	return &v2rayclient.VmessOutbound{
		Address: addr,
		AlterID: uint32(alterID),
		ID:      username,
		Port:    uint32(port),
	}, nil
}

// AsClient runs adapter in clients mode.
func AsClient() {
	conf := new(clientConfig)
	readConfigFile(conf)

	sesscl := newProductSessClient(conf.Sess)

	changesChan := connChangeSubscribe(sesscl)

	conn := newV2RayAPIConn(conf.V2Ray.API)

	statsclient := newV2RayStatsClient(conn, conf.V2Ray.InboundTag)

	mon := newMonitor(statsclient, conf.Monitor)

	configurer := v2rayclient.NewConfigurer(conn)

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
			req, err := newConfigureRequest(username, endpoint.AdditionalParams)
			if err != nil {
				// TODO: log warning or fatal.
				continue
			}
			configurer.ConfigureVmess(context.Background(), req)
			// 2. Start monitoring.
			mon.Start(username)
			// TODO: 3. Start reading v2ray logs to detect and handle connection drops.
			// ? How to recognize logs particularly for this connection ?
		case sess.ConnStop:
			// 1. Stop or restore v2ray configuration.
			configurer.RemoveVmess(context.Background())
			// TODO: 2. Stop reading v2ray logs for this connection.
			// 3. Stop monitoring.
			mon.Stop(username)
		}
	}
}
