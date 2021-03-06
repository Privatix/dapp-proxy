package adapter

import (
	"io"

	"google.golang.org/grpc"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	"github.com/privatix/dapp-proxy/plugin/v2ray-client"
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
			err := adapterSessClient.UpdateSession(report.Channel, report.Usage, report.Last)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

func runAdapter(conf *Config, beforeStart func(), onConnCreate, onConnStart, onConnStop func(*data.Endpoint, *sess.ConnChangeResult)) {
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

	beforeStart()

	go handleReports()

	adapterLogger.Info("Starting proxy adapter")

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
		case sess.ConnCreate:
			onConnCreate(endpoint, change)
		case sess.ConnStart:
			onConnStart(endpoint, change)
		case sess.ConnStop:
			onConnStop(endpoint, change)
		}
	}
}
