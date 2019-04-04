package adapter

import (
	"context"

	"github.com/privatix/dappctrl/sess"

	"github.com/privatix/dappctrl/data"
)

// AsClient runs adapter in client mode.
func AsClient(conf *Config) {
	onConnCreate := func(_ *data.Endpoint, _ *sess.ConnChangeResult) {}
	onConnStart := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("configuring proxy to connect")
		req, err := newConfigureRequest(*endpoint.Username, endpoint.AdditionalParams)
		if err != nil {
			adapterLogger.Warn("could not build request to configure")
			return
		}
		adapterLogger.Add("adapterConfigurerequest", *req).Debug("configure vmess request")
		adapterConfigurer.ConfigureVmess(context.Background(), req)
		adapterMon.Start(*endpoint.Username, change.Channel)
		// TODO: Start reading v2ray logs to detect and handle connection drops.
		// ? How to recognize logs particularly for this connection ?
	}
	onConnStop := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("removing proxy configuration to connect")
		adapterConfigurer.RemoveVmess(context.Background())
		// TODO: Stop reading v2ray logs for this connection.
		adapterMon.Stop(*endpoint.Username, change.Channel)
	}

	runAdapter(conf, func() {}, onConnCreate, onConnStart, onConnStop)
}
