package monitor_test

import (
	"testing"
	"time"

	"github.com/privatix/dapp-proxy/plugin/monitor"
	"github.com/privatix/dappctrl/util/log"
)

type testUsageGetter struct {
	usage uint64
}

func newTestUsageGetter() *testUsageGetter {
	return &testUsageGetter{}
}

func (g *testUsageGetter) Get() (uint64, error) {
	return g.usage, nil
}

func TestMonitor(t *testing.T) {
	tUsageGetter := newTestUsageGetter()

	logger, err := log.NewTestLogger(nil, false)
	if err != nil {
		panic(err)
	}
	mon := monitor.NewMonitor(time.Millisecond, logger)

	tUsageGetter.usage = 100

	go mon.Start("bar", tUsageGetter)

	select {
	case <-time.After(time.Second):
		t.Fatal("usage was not reported: timeout")
	case v := <-mon.Reports:
		if v.Channel != "bar" || v.Usage != 100 || !v.First || v.Last {
			t.Fatalf("unexpected usage reported: %+v", v)
		}
	}

	select {
	case <-time.After(time.Second):
		t.Fatal("usage was not reported: timeout")
	case v := <-mon.Reports:
		if v.Channel != "bar" || v.Usage != 100 || v.First || v.Last {
			t.Fatalf("unexpected usage reported: %+v", v)
		}
	}

	mon.Stop("bar")

	works := false
	for v := range mon.Reports {
		works = v.Channel == "bar" && v.Usage == 100 && !v.First && v.Last
		if works {
			return
		}
	}
	t.Fatal("last report was not sent")
}
