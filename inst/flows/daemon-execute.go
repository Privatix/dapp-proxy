package flows

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type daemonExecute struct {
	execPath string
	confPath string

	proc *os.Process
}

// Start non-blocking starts v2ray.
func (e *daemonExecute) Start() {
	e.runInBackground()
}

// Stop non-blocking stops v2ray.
func (e *daemonExecute) Stop() {
	if e.proc != nil {
		err := e.proc.Kill()
		if err == nil {
			e.proc = nil
		}
	}
}

// Run blocking runs v2ray.
func (e *daemonExecute) Run() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close(interrupt)

	e.runInBackground()

	for {
		select {
		case <-interrupt:
			if e.proc != nil {
				e.proc.Kill()
			}
			break
		}
	}
}

func (e *daemonExecute) runInBackground() {
	go func() {
		if err := e.run(); err != nil {
			os.Exit(2)
		}
		os.Exit(0)
	}()
}

func (e *daemonExecute) run() error {
	cmd := exec.Command(e.execPath, "--config", e.confPath)

	err := cmd.Start()
	if err != nil {
		return err
	}

	e.proc = cmd.Process

	return cmd.Wait()
}
