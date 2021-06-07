package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xing/kubernetes-oom-event-generator/src/controller"
	"github.com/xing/kubernetes-oom-event-generator/src/util"
)

var opts struct {
	Verbose int  `env:"VERBOSE" short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool `long:"version" description:"Print version information"`
}

// VERSION represents the current version of the release.
const VERSION = "v1.1.0"

func main() {
	util.ParseArgs(&opts)

	if opts.Version {
		printVersion()
		return
	}

	stopChan := make(chan struct{})
	util.InstallSignalHandler(stopChan)

	eventGenerator := controller.NewController(stopChan)
	util.InstallSignalHandler(eventGenerator.Stop)

	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf("0.0.0.0:10254")
	go func() { glog.Fatal(http.ListenAndServe(addr, nil)) }()

	err := eventGenerator.Run()
	if err != nil {
		glog.Fatal(err)
	}
}

func printVersion() {
	fmt.Printf("kubernetes-oom-event-generator %s %s/%s %s\n", VERSION, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
