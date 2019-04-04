package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/privatix/dappctrl/util/log"
)

type action int

// UsageGetter abstracts how traffic usage is computed/extracted.
type UsageGetter interface {
	Get(username string) (uint64, error)
}

// Report is usage report format.
type Report struct {
	Channel string
	Usage   uint64
	First   bool
	Last    bool
}

// Monitor counts and reports traffic usage.
type Monitor struct {
	Reports chan *Report

	mu            sync.Mutex
	logger        log.Logger
	cancel        map[string]context.CancelFunc
	reportsPeriod time.Duration
	usage         UsageGetter
}

// NewMonitor creates a monitor.
func NewMonitor(usage UsageGetter, period time.Duration, logger log.Logger) *Monitor {
	return &Monitor{
		Reports:       make(chan *Report),
		logger:        logger.Add("type", "Monitor"),
		cancel:        make(map[string]context.CancelFunc),
		reportsPeriod: period,
		usage:         usage,
	}
}

// Start start monitor traffic usage for username.
func (m *Monitor) Start(username, channel string) {
	logger := m.logger.Add("method", "Start", "username", username, "channel", channel)
	logger.Info("start monitoring")

	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.cancel[username+channel] = cancel
	m.mu.Unlock()

	go func() {
		ticker := time.NewTicker(m.reportsPeriod)
		defer ticker.Stop()

		first := true

		for {
			select {
			case <-ticker.C:
				m.reportUsage(username, channel, first, false)
				first = false
			case <-ctx.Done():
				m.reportUsage(username, channel, false, true)
			}
		}
	}()
}

// Stop stop monitoring traffic usage for username.
func (m *Monitor) Stop(username, channel string) {
	logger := m.logger.Add("username", username)

	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.cancel[username+channel]
	if !ok {
		logger.Warn("stop request for not monitoring username")
		return
	}

	cancel()
	delete(m.cancel, username)
}

func (m *Monitor) reportUsage(username, channel string, f, l bool) {
	logger := m.logger.Add("username", username)

	logger.Debug("reporting usage")

	usage, err := m.usage.Get(username)
	if err != nil {
		logger.Error(err.Error())
	}

	m.Reports <- &Report{
		Channel: channel,
		Usage:   usage,
		First:   f,
		Last:    l,
	}
}
