package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/privatix/dappctrl/util/log"
)

// UsageGetter abstracts how traffic usage is computed/extracted.
type UsageGetter interface {
	Get() (uint64, error)
}

// Report is usage report format.
type Report struct {
	Channel string
	Usage   uint64
	First   bool
}

// Monitor counts and reports traffic usage.
type Monitor struct {
	Reports chan *Report

	mu            sync.Mutex
	logger        log.Logger
	cancel        map[string]context.CancelFunc
	reportsPeriod time.Duration
	usages        map[string]UsageGetter
}

// NewMonitor creates a monitor.
func NewMonitor(period time.Duration, logger log.Logger) *Monitor {
	return &Monitor{
		Reports:       make(chan *Report),
		logger:        logger.Add("type", "Monitor"),
		cancel:        make(map[string]context.CancelFunc),
		reportsPeriod: period,
		usages:        make(map[string]UsageGetter),
	}
}

// Start start monitor traffic usage for username.
func (m *Monitor) Start(channel string, getter UsageGetter) {
	logger := m.logger.Add("method", "Start", "channel", channel)
	logger.Info("start monitoring")

	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.usages[channel] = getter
	m.cancel[channel] = cancel
	m.mu.Unlock()

	go func() {
		ticker := time.NewTicker(m.reportsPeriod)
		defer ticker.Stop()

		first := true

		for {
			select {
			case <-ticker.C:
				m.reportUsage(channel, first)
				first = false
			case <-ctx.Done():
				m.mu.Lock()
				delete(m.usages, channel)
				m.mu.Unlock()
				return
			}
		}
	}()
}

// Stop stop monitoring traffic usage for username.
func (m *Monitor) Stop(channel string) {
	logger := m.logger.Add("channel", channel)

	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.cancel[channel]
	if !ok {
		logger.Warn("stop request for not monitoring username")
		return
	}

	cancel()
	delete(m.cancel, channel)
}

func (m *Monitor) reportUsage(channel string, first bool) {
	logger := m.logger.Add()

	logger.Debug("reporting usage")

	usage, err := m.usages[channel].Get()
	if err != nil {
		logger.Error(err.Error())
	}

	m.Reports <- &Report{
		Channel: channel,
		Usage:   usage,
		First:   first,
	}
}
