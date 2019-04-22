package adapter

import (
	"context"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	"github.com/privatix/dapp-proxy/plugin/v2ray-client"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
)

// AsClient runs adapter in client mode.
func AsClient(conf *Config) {
	onConnCreate := func(_ *data.Endpoint, _ *sess.ConnChangeResult) {}
	onConnStart := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("configuring proxy to connect")
		req, err := newConfigureRequest(*endpoint.Username, endpoint.AdditionalParams)
		must("could not build request to configure", err)
		adapterLogger.Add("adapterConfigurerequest", *req).Debug("configure vmess request")
		adapterConfigurer.ConfigureVmess(context.Background(), req)
		u := v2rayclient.NewInboundUsageGetter(adapterV2RayConn, conf.V2Ray.InboundTag)
		adapterMon.Start(change.Channel, monitor.NewUsageGetterAdapter(u))
		// TODO: Start reading v2ray logs to detect and handle connection drops.
		// ? How to recognize logs particularly for this connection ?
	}
	onConnStop := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("removing proxy configuration to connect")
		adapterConfigurer.RemoveVmess(context.Background())
		// TODO: Stop reading v2ray logs for this connection.
		adapterMon.Stop(change.Channel)
	}

	runAdapter(conf, func() {}, onConnCreate, onConnStart, onConnStop)
}
