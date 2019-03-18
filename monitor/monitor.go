package monitor

import (
	"time"
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
	username string
	action   action
}

// Report is usage report format.
type Report struct {
	Username string
	Usage    uint64
	First    bool
	Last     bool
}

// Monitor counts and reports traffic usage.
type Monitor struct {
	Reports chan *Report

	commands      map[string]chan *command
	reportsPeriod time.Duration
	usage         UsageGetter
}

// NewMonitor creates a monitor.
func NewMonitor(usage UsageGetter, period time.Duration) *Monitor {
	return &Monitor{
		Reports:       make(chan *Report),
		commands:      make(map[string]chan *command),
		reportsPeriod: period,
		usage:         usage,
	}
}

// Start start monitor traffic usage for username.
func (m *Monitor) Start(username string) {
	ch := make(chan *command)
	m.commands[username] = ch
	go func() {
		for cmd := range ch {
			m.reportUsage(cmd)

			go func(username string, act action) {
				ch := m.commands[username]
				if act == stopMonitoring {
					delete(m.commands, username)
					close(ch)
					return
				}
				ch <- &command{
					username: username,
					action:   monitoring,
				}
			}(cmd.username, cmd.action)
		}
	}()
	ch <- &command{
		username: username,
		action:   startMonitoring,
	}
}

// Stop stop monitoring traffic usage for username.
func (m *Monitor) Stop(username string) {
	ch, ok := m.commands[username]
	if !ok {
		// TODO: log warning
		return
	}
	go func() {
		ch <- &command{
			username: username,
			action:   stopMonitoring,
		}
	}()
}

func (m *Monitor) reportUsage(cmd *command) {
	usage, err := m.usage.Get(cmd.username)
	if err != nil {
		// TODO: log error or fatal.
		return
	}

	m.Reports <- &Report{
		Username: cmd.username,
		Usage:    usage,
		First:    cmd.action == startMonitoring,
		Last:     cmd.action == stopMonitoring,
	}
}
