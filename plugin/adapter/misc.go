package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	v2rayclient "github.com/privatix/dapp-proxy/plugin/v2ray-client"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util/log"
)

func must(msg string, err error) {
	if err != nil {
		if msg != "" {
			adapterLogger.Fatal(msg + ": " + err.Error())
		}
		adapterLogger.Fatal(err.Error())
	}
}

func createLogger(conf *log.FileConfig) (log.Logger, io.Closer) {
	elog, err := log.NewStderrLogger(conf.WriterConfig)
	must("", err)

	flog, closer, err := log.NewFileLogger(conf)
	must("", err)

	logger := log.NewMultiLogger(elog, flog)

	return logger, closer
}

func newV2RayUsersClient(conn *grpc.ClientConn, inboundTag string, alterID uint32) *v2rayclient.UsersClient {
	client := v2rayclient.NewUsersClient(conn, inboundTag, alterID)
	return client
}

func newV2RayAPIConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	must("could not dial v2ray api", err)
	return conn
}

func newProductSessClient(conf SessConfig) *sess.Client {
	client, err := sess.Dial(context.Background(), conf.Endpoint,
		conf.Origin, conf.Product, conf.Password)
	must("", err)
	return client
}

func connChangeSubscribe(c *sess.Client) chan *sess.ConnChangeResult {
	ret := make(chan *sess.ConnChangeResult)
	subn, err := c.ConnChange(ret)
	must("", err)

	go func() {
		select {
		case err := <-subn.Err():
			must("unexpected end of subscription to connection changes", err)
		}
	}()

	return ret
}

func newMonitor(conf MonitorConfig) *monitor.Monitor {
	return monitor.NewMonitor(
		time.Duration(conf.CountPeriod)*time.Second,
		adapterLogger,
	)
}

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
