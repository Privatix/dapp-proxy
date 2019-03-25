package adapter

import (
	"context"
	"flag"
	"time"

	"google.golang.org/grpc"

	"github.com/privatix/dapp-proxy/monitor"
	v2rayclient "github.com/privatix/dapp-proxy/v2ray-client"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

func must(msg string, err error) {
	if err != nil {
		if msg != "" {
			panic(msg + ": " + err.Error())
		}
		panic(err)
	}
}

func readConfigFile(conf interface{}) string {
	fconfig := flag.String(
		"config", "config.json", "Configuration file")
	flag.Parse()

	err := util.ReadJSONFile(*fconfig, &conf)
	must("failed to read configuration: ", err)
	return *fconfig
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

func newV2RayStatsClient(conn *grpc.ClientConn, inboundTag string) *v2rayclient.StatsClient {
	client := v2rayclient.NewStatsClient(conn, inboundTag)
	return client
}

func newProductSessClient(conf sessConfig) *sess.Client {
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

func newMonitor(client *v2rayclient.StatsClient, conf monitorConfig) *monitor.Monitor {
	return monitor.NewMonitor(
		monitor.NewV2RayClientUsageGetter(client),
		time.Duration(conf.CountPeriod)*time.Second)
}

func handleReports(mon *monitor.Monitor, sesscl *sess.Client) {
	for report := range mon.Reports {
		if report.First {
			_, err := sesscl.StartSession("", report.Username, 0)
			if err != nil {
				// TODO: log error or fatal.
			}
		}
		err := sesscl.UpdateSession(report.Username, report.Usage, report.Last)
		if err != nil {
			// TODO: log error or fatal.
		}
	}
}
