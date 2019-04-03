package monitor

import (
	"sync"
	"time"

	"github.com/privatix/dappctrl/util/log"
)

type action int

// Actions
const (
	startMonitoring action = iota
	stopMonitoring
	monitoring
)

// UsageGetter abstracts how traffic usage is computed/extracted.
type UsageGetter interface {
	Get(username string) (uint64, error)
}

// command is request to start or stop monitoring for username.
type command struct {
	channel  string
	username string
	action   action
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

	commands      map[string]chan *command
	logger        log.Logger
	mu            sync.Mutex
	reportsPeriod time.Duration
	usage         UsageGetter
}

// NewMonitor creates a monitor.
func NewMonitor(usage UsageGetter, period time.Duration, logger log.Logger) *Monitor {
	return &Monitor{
		Reports:       make(chan *Report),
		commands:      make(map[string]chan *command),
		logger:        logger.Add("type", "Monitor"),
		reportsPeriod: period,
		usage:         usage,
	}
}

// Start start monitor traffic usage for username.
func (m *Monitor) Start(username, channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := m.logger.Add("method", "Start", "username", username, "channel", channel)
	logger.Info("start monitoring")

	ch := make(chan *command)
	m.commands[username] = ch
	go func() {
		for cmd := range ch {
			logger.Add("cmd", *cmd).Debug("recieved command")
			m.reportUsage(cmd)

			go func(username string, act action) {
				m.mu.Lock()
				defer m.mu.Unlock()
				ch := m.commands[username]
				if act == stopMonitoring {
					delete(m.commands, username)
					close(ch)
					return
				}
				time.Sleep(m.reportsPeriod)
				ch <- &command{
					channel:  channel,
					username: username,
					action:   monitoring,
				}
			}(cmd.username, cmd.action)
		}
	}()
	ch <- &command{
		channel:  channel,
		username: username,
		action:   startMonitoring,
	}
}

// Stop stop monitoring traffic usage for username.
func (m *Monitor) Stop(username, channel string) {
	logger := m.logger.Add("username", username)

	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		ch, ok := m.commands[username]
		if !ok {
			logger.Warn("stop request for not monitoring username")
			return
		}
		ch <- &command{
			channel:  channel,
			username: username,
			action:   stopMonitoring,
		}
	}()
}

func (m *Monitor) reportUsage(cmd *command) {
	logger := m.logger.Add("command", *cmd)
	logger.Debug("reporting usage")

	usage, err := m.usage.Get(cmd.username)
	if err != nil {
		logger.Error(err.Error())
	}

	m.Reports <- &Report{
		Channel: cmd.channel,
		Usage:   usage,
		First:   cmd.action == startMonitoring,
		Last:    cmd.action == stopMonitoring,
	}
}
