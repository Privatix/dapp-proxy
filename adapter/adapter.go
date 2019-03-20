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

func readConfigFile(conf interface{}) {
	fconfig := flag.String(
		"config", "config.json", "Configuration file")
	flag.Parse()

	err := util.ReadJSONFile(*fconfig, &conf)
	must("failed to read configuration: ", err)
}

func newV2RayClient(addr, inboundTag string, alterID uint32) *v2rayclient.Client {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	must("could not dial v2ray api", err)
	client := v2rayclient.NewClient(conn, inboundTag, alterID)
	return client
}

func dialSess(conf sessConfig) *sess.Client {
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

func newMonitor(client *v2rayclient.Client, conf monitorConfig) *monitor.Monitor {
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