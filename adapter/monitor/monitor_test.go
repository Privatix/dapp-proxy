package monitor_test

import (
	"testing"
	"time"

	"github.com/privatix/dapp-proxy/adapter/monitor"
	"github.com/privatix/dappctrl/util/log"
)

type testUsageGetter struct {
	reports map[string]uint64
}

func newTestUsageGetter() *testUsageGetter {
	return &testUsageGetter{make(map[string]uint64)}
}

func (usage *testUsageGetter) Get(user string) (uint64, error) {
	return usage.reports[user], nil
}

func TestMonitor(t *testing.T) {
	usage := newTestUsageGetter()

	logger, err := log.NewTestLogger(nil, false)
	if err != nil {
		panic(err)
	}
	mon := monitor.NewMonitor(usage, time.Millisecond, logger)

	usage.reports["foo"] = 100

	go mon.Start("foo", "bar")

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

	mon.Stop("foo", "bar")

	works := false
	for v := range mon.Reports {
		works = v.Channel == "bar" && v.Usage == 100 && !v.First && v.Last
		if works {
			return
		}
	}
	t.Fatal("last report was not sent")
}
