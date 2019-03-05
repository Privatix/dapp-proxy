package main

import (
	"context"
	"flag"
	"fmt"

	client "github.com/privatix/dapp-proxy/v2ray-client"
	"github.com/privatix/dappctrl/util"
)

type config struct {
	AlterID    uint32
	API        string
	InboundTag string
}

func readConfig() *config {
	fconfig := flag.String(
		"config", "config.json", "Configuration file")
	flag.Parse()

	conf := new(config)
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s\n", err))
	}
	return conf
}

func main() {
	conf := readConfig()
	client, err := client.NewClient(conf.API, conf.InboundTag, conf.AlterID)
	if err != nil {
		panic(err)
	}
	client.AddUser(context.Background(), "b831381d-6324-4d53-ad4f-8cda48b30811")
}
