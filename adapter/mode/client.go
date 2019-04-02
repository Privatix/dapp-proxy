package mode

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	v2rayclient "github.com/privatix/dapp-proxy/adapter/v2ray-client"
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

	port, err := strconv.ParseUint(portRaw, 10, 64)
	if err != nil {
		return nil, err
	}

	alterIDRaw, ok := prodconf[productAlterID]
	if !ok {
		return nil, fmt.Errorf("could not find %s", productAlterID)
	}

	alterID, err := strconv.ParseUint(alterIDRaw, 10, 64)
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
func AsClient(conf *ClientConfig) {

	sesscl := newProductSessClient(conf.Sess)

	changesChan := connChangeSubscribe(sesscl)

	conn := newV2RayAPIConn(conf.V2Ray.API)

	statsclient := newV2RayStatsClient(conn, conf.V2Ray.InboundTag)

	logger, closer := createLogger(conf.FileLog)
	defer closer.Close()

	mon := newMonitor(statsclient, conf.Monitor, logger)

	configurer := v2rayclient.NewConfigurer(conn)

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
		case sess.ConnStart:
			req, err := newConfigureRequest(username, endpoint.AdditionalParams)
			if err != nil {
				logger.Warn("could not build request to configure")
				continue
			}
			configurer.ConfigureVmess(context.Background(), req)
			mon.Start(username, change.Channel)
			// TODO: Start reading v2ray logs to detect and handle connection drops.
			// ? How to recognize logs particularly for this connection ?
		case sess.ConnStop:
			configurer.RemoveVmess(context.Background())
			// TODO: Stop reading v2ray logs for this connection.
			mon.Stop(username, change.Channel)
		}
	}
}
