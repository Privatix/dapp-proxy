package adapter

import (
	"context"
	"flag"
	"time"

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

func readConfigFile() *config {
	fconfig := flag.String(
		"config", "config.json", "Configuration file")
	flag.Parse()

	conf := new(config)
	err := util.ReadJSONFile(*fconfig, &conf)
	must("failed to read configuration: ", err)
	return conf
}

func dialV2Ray(conf v2rayConfig) *v2rayclient.Client {
	client, err := v2rayclient.Dial(conf.API, conf.InboundTag, conf.AlterID)
	must("", err)
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
