package mode

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"

	"github.com/privatix/dapp-proxy/adapter/monitor"
	v2rayclient "github.com/privatix/dapp-proxy/adapter/v2ray-client"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util/log"
)

func must(msg string, err error) {
	if err != nil {
		if msg != "" {
			panic(msg + ": " + err.Error())
		}
		panic(err)
	}
}

func createLogger(conf *log.FileConfig) (log.Logger, io.Closer) {
	elog, err := log.NewStderrLogger(conf.WriterConfig)
	must("", err)

	flog, closer, err := log.NewFileLogger(conf)
	must("", err)

	logger := log.NewMultiLogger(elog, flog)

	return logger, closer
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

func newProductSessClient(conf SessConfig) *sess.Client {
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

func newMonitor(client *v2rayclient.StatsClient, conf MonitorConfig, logger log.Logger) *monitor.Monitor {
	return monitor.NewMonitor(
		monitor.NewV2RayClientUsageGetter(client),
		time.Duration(conf.CountPeriod)*time.Second,
		logger,
	)
}

func handleReports(mon *monitor.Monitor, sesscl *sess.Client, logger log.Logger) {
	logger = logger.Add("method", "handleReports")

	for report := range mon.Reports {
		logger = logger.Add("report", *report)

		if report.First {
			_, err := sesscl.StartSession("", report.Channel, 0)
			if err != nil {
				logger.Fatal(err.Error())
			}
		} else {
			err := sesscl.UpdateSession(report.Channel, report.Usage, report.Last)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}
