package monitor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/privatix/dapp-proxy/monitor"
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

	mon := monitor.NewMonitor(usage, time.Millisecond)

	usage.reports["foo"] = 100

	go mon.Start("foo")

	select {
	case <-time.After(time.Second):
		t.Fatal("usage was not reported: timeout")
	case v := <-mon.Reports:
		if v.Username != "foo" || v.Usage != 100 || !v.First || v.Last {
			t.Fatalf("unexpected usage reported: %+v", v)
		}
	}

	select {
	case <-time.After(time.Second):
		t.Fatal("usage was not reported: timeout")
	case v := <-mon.Reports:
		if v.Username != "foo" || v.Usage != 100 || v.First || v.Last {
			t.Fatalf("unexpected usage reported: %+v", v)
		}
	}

	mon.Stop("foo")

	works := false
	for v := range mon.Reports {
		fmt.Println(v)
		works = v.Username == "foo" && v.Usage == 100 && !v.First && v.Last
		if works {
			return
		}
	}
	t.Fatal("last report was not sent")
}
