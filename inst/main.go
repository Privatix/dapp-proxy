package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/flow"
	"github.com/privatix/dapp-proxy/inst/flows"
	"github.com/privatix/dappctrl/util/log"
)

func createLogger() (log.Logger, io.Closer) {
	failOnErr := func(err error) {
		if err != nil {
			panic("failed to create a logger: " + err.Error())
		}
	}

	elog, err := log.NewStderrLogger(log.NewWriterConfig())
	failOnErr(err)

	f := flag.NewFlagSet("", flag.ContinueOnError)
	p := f.String("proddir", "..", "Product install directory")

	if len(os.Args) > 2 && !strings.EqualFold(os.Args[1], "install") {
		f.Parse(os.Args[2:])
	}

	if strings.EqualFold(*p, "..") {
		*p = filepath.Join(filepath.Dir(os.Args[0]), *p)
	}

	path, _ := filepath.Abs(*p)
	path = filepath.ToSlash(path)

	fileName := filepath.Join(path, "log/installer-%Y-%m-%d.log")

	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     fileName,
		FileMode:     0644,
	}

	flog, closer, err := log.NewFileLogger(logConfig)
	failOnErr(err)

	return log.NewMultiLogger(elog, flog), closer
}

func main() {
	logger, closer := createLogger()
	defer closer.Close()

	if len(os.Args) <= 1 {
		fmt.Println(flows.RootHelp)
		return
	}

	arg := os.Args[1]

	ok, _ := flow.Execute(logger, arg, map[string]flow.Flow{
		flows.MethodInstall:   flows.Install(),
		flows.MethodUpdate:    flows.Update(),
		flows.MethodStart:     flows.Start(),
		flows.MethodStop:      flows.Stop(),
		flows.MethodRemove:    flows.Remove(),
		flows.MethodRunV2Ray:  flows.RunV2Ray(),
		flows.MethodRunPlugin: flows.RunPlugin(),
	}, flows.NewProxyInstallation())

	if !ok {
		fmt.Println(flows.RootHelp)
	}
}
