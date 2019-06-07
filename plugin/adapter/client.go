package adapter

import (
	"context"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	"github.com/privatix/dapp-proxy/plugin/osconnector"
	v2rayclient "github.com/privatix/dapp-proxy/plugin/v2ray-client"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
)

// AsClient runs adapter in client mode.
func AsClient(conf *Config) {
	onConnStart := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		var err error
		adapterLogger.Info("configuring proxy to connect")
		req, err := newConfigureRequest(*endpoint.Username, endpoint.AdditionalParams)
		must("could not build request to configure", err)
		adapterLogger.Add("adapterConfigurerequest", *req).Debug("configure vmess request")

		err = adapterConfigurer.ConfigureVmess(context.Background(), req)
		if err != nil {
			adapterConfigurer.RemoveVmess(context.Background())
			adapterLogger.Fatal("could not configure vmess: " + err.Error())
		}

		adapterLogger.Info("configuring operating system to use proxy")
		err = osconnector.ConfigureWithScript(conf.ConfigureProxyScript, conf.ProxyBackupFile, conf.ProxyPort)
		must("could not configure operating system to use proxy", err)

		u := v2rayclient.NewInboundUsageGetter(adapterV2RayConn, conf.V2Ray.InboundTag)
		adapterLogger.Info("requesting traffic counter reset")
		err = u.RequestReset(context.Background())
		must("", err)

		adapterMon.Start(change.Channel, monitor.NewUsageGetterAdapter(u))
		// TODO: Start reading v2ray logs to detect and handle connection drops.
		// ? How to recognize logs particularly for this connection ?
	}

	onConnStop := func(endpoint *data.Endpoint, change *sess.ConnChangeResult) {
		adapterLogger.Info("removing proxy configuration to connect")

		err := adapterConfigurer.RemoveVmess(context.Background())
		must("could not remove vmess", err)

		adapterLogger.Info("configuring operating system to stop using proxy")
		err = osconnector.RollbackWithScript(conf.ConfigureProxyScript, conf.ProxyBackupFile)
		must("could not configure operating system to stop using proxy", err)
		adapterSessClient.StopSession(change.Channel)
		must("failed to stop session", err)

		// TODO: Stop reading v2ray logs for this connection.
		adapterMon.Stop(change.Channel)
	}

	runAdapter(conf, func() {}, onConnStart, onConnStop)
}
