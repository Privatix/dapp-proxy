package monitor

import (
	"time"
)

type action bool

// Actions
const (
	StartMonitoring action = true
	StopMonitoring  action = false
)

// UsageGetter abstracts how traffic usage is computed/extracted.
type UsageGetter interface {
	Get(username string) (uint, error)
}

// Command is request to start or stop monitoring for username.
type Command struct {
	Username string
	Action   action
}

// Report is usage report format.
type Report struct {
	Username string
	Usage    uint
}

// Monitor counts and reports traffic usage.
type Monitor struct {
	Commands chan *Command
	Reports  chan *Report

	reportsPeriod time.Duration
	usernames     map[string]bool
	usage         UsageGetter
}

// NewMonitor creates a monitor.
func NewMonitor(usage UsageGetter, period time.Duration) *Monitor {
	return &Monitor{
		Commands:      make(chan *Command),
		Reports:       make(chan *Report),
		reportsPeriod: period,
		usernames:     make(map[string]bool),
		usage:         usage,
	}
}

// Start start monitor routine.
func (m *Monitor) Start() {

	go func() {
		for range time.Tick(m.reportsPeriod) {
			m.reportUsages()
		}
	}()

	for cmd := range m.Commands {
		switch cmd.Action {
		case StartMonitoring:
			if _, ok := m.usernames[cmd.Username]; ok {
				// TODO: log warning.
				continue
			}
			m.usernames[cmd.Username] = true
		case StopMonitoring:
			delete(m.usernames, cmd.Username)
		}
	}
}

func (m *Monitor) reportUsages() {
	for username := range m.usernames {
		usage, err := m.usage.Get(username)
		if err != nil {
			// TODO: log error or fatal.
			return
		}

		report := &Report{
			Username: username,
			Usage:    usage,
		}
		go func() {
			m.Reports <- report
		}()
	}
}
