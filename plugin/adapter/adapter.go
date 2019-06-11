package adapter

import (
	"io"

	"google.golang.org/grpc"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	v2rayclient "github.com/privatix/dapp-proxy/plugin/v2ray-client"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util/log"
)

var (
	adapterV2RayConn   *grpc.ClientConn
	adapterChangesChan chan *sess.ConnChangeResult
	adapterConfigurer  *v2rayclient.Configurer
	adapterLogger      log.Logger
	adapterMon         *monitor.Monitor
	adapterSessClient  *sess.Client
	adapterUsersClient *v2rayclient.UsersClient
)

func handleReports() {
	logger := adapterLogger.Add("method", "handleReports")

	logger.Debug("start handling usage reports")

	for report := range adapterMon.Reports {
		logger = logger.Add("report", *report)

		if report.First {
			_, err := adapterSessClient.StartSession("", report.Channel, 0)
			if err != nil {
				logger.Fatal(err.Error())
			}
		} else {
			err := adapterSessClient.UpdateSession(report.Channel, report.Usage)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

func runAdapter(conf *Config, beforeStart func(), onConnStart, onConnStop func(*data.Endpoint, *sess.ConnChangeResult)) {
	adapterSessClient = newProductSessClient(conf.Sess)

	adapterV2RayConn = newV2RayAPIConn(conf.V2Ray.API)

	adapterConfigurer = v2rayclient.NewConfigurer(adapterV2RayConn)

	adapterUsersClient = newV2RayUsersClient(adapterV2RayConn, conf.V2Ray.InboundTag,
		conf.V2Ray.AlterID)

	adapterChangesChan = connChangeSubscribe(adapterSessClient)

	var closer io.Closer
	adapterLogger, closer = createLogger(conf.FileLog)
	defer closer.Close()

	adapterMon = newMonitor(conf.Monitor)

	channelStorage := newActiveChannelStorage(conf.ChannelDir)

	beforeStart()

	adapterLogger.Info("Starting proxy adapter")

	if ch, err := channelStorage.load(); err != nil {
		adapterLogger.Fatal(err.Error())
	} else if ch != "" {
		adapterLogger.Info("Stop session for left over channel: " + ch)
		err := adapterSessClient.StopSession(ch)
		if err != nil {
			adapterLogger.Fatal(err.Error())
		}
		if err := channelStorage.remove(); err != nil {
			adapterLogger.Fatal(err.Error())
		}
	} else {
		adapterLogger.Info("No left over channel")
	}

	go handleReports()

	for change := range adapterChangesChan {
		logger := adapterLogger.Add("connectionChange", *change)

		logger.Debug("received connection change")

		endpoint, err := adapterSessClient.GetEndpoint(change.Channel)
		if err != nil {
			logger.Fatal(err.Error())
		}

		logger = logger.Add("endpoint", *endpoint)

		if endpoint.Username == nil {
			logger.Fatal("username of connection change is empty")
		}

		switch change.Status {
		case sess.ConnStart:
			if err := channelStorage.store(change.Channel); err != nil {
				logger.Fatal(err.Error())
			}
			onConnStart(endpoint, change)
		case sess.ConnStop:
			onConnStop(endpoint, change)
			if err := channelStorage.remove(); err != nil {
				logger.Fatal(err.Error())
			}
		}
	}
}
