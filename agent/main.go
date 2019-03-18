package main

import (
	"context"
	"flag"
	"time"

	"github.com/privatix/dapp-proxy/monitor"
	v2rayclient "github.com/privatix/dapp-proxy/v2ray-client"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

type config struct {
	V2Ray   v2rayConf
	Sess    sessConfig
	Monitor monitorConfig
}

type v2rayConf struct {
	AlterID    uint32
	API        string
	InboundTag string
}

type sessConfig struct {
	Endpoint string
	Origin   string
	Product  string
	Password string
}

type monitorConfig struct {
	CountPeriod uint // in seconds.
}

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

func dialV2Ray(conf v2rayConf) *v2rayclient.Client {
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

func main() {
	conf := readConfigFile()

	client := dialV2Ray(conf.V2Ray)

	sesscl := dialSess(conf.Sess)

	changesChan := connChangeSubscribe(sesscl)

	mon := newMonitor(client, conf.Monitor)

	go mon.Start()

	for change := range changesChan {
		endpoint, err := sesscl.GetEndpoint(change.Channel)
		must("", err)

		if endpoint.Username == nil {
			// TODO: log error or fatal.
		}

		switch change.Status {
		case sess.ConnStart:
			err = client.AddUser(context.Background(), *endpoint.Username)
			must("", err)
			mon.Commands <- &monitor.Command{
				Username: *endpoint.Username,
				Action:   monitor.StartMonitoring,
			}
		case sess.ConnStop:
			err = client.RemoveUser(context.Background(), *endpoint.Username)
			must("", err)
			mon.Commands <- &monitor.Command{
				Username: *endpoint.Username,
				Action:   monitor.StopMonitoring,
			}
		}
	}
}
