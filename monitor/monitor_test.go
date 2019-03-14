package monitor_test

import (
	"testing"
	"time"

	"github.com/privatix/dapp-proxy/monitor"
)

type testUsageGetter struct {
	reports map[string]uint
}

func newTestUsageGetter() *testUsageGetter {
	return &testUsageGetter{make(map[string]uint)}
}

func (usage *testUsageGetter) Get(user string) (uint, error) {
	return usage.reports[user], nil
}

func TestMonitor(t *testing.T) {
	usage := newTestUsageGetter()

	mon := monitor.NewMonitor(usage, time.Millisecond)
	go mon.Start()

	usage.reports["foo"] = 100

	go func() {
		mon.Commands <- &monitor.Command{
			Username: "foo",
			Action:   monitor.StartMonitoring,
		}
	}()

	select {
	case <-time.After(time.Second):
		t.Fatal("usage was not reported: timeout")
	case v := <-mon.Reports:
		if v.Username != "foo" || v.Usage != 100 {
			t.Fatalf("unexpected usage reported: %+v", v)
		}
	}
}
