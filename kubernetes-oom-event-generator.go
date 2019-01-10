package main

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xing/kubernetes-oom-event-generator/src/util"
	"github.com/xing/kubernetes-oom-event-generator/src/controller"
)

var opts struct {
	Verbose int `env:"VERBOSE" long:"verbose" description:"Show verbose debug information"`
}

func main() {
	util.ParseArgs(&opts)

	stopChan := make(chan struct{})
	util.InstallSignalHandler(stopChan)

	eventGenerator := controller.NewController(stopChan)
	util.InstallSignalHandler(eventGenerator.Stop)

	http.Handle("/metrics", prometheus.Handler())
	addr := fmt.Sprintf("0.0.0.0:10254")
	go func() { glog.Fatal(http.ListenAndServe(addr, nil)) }()

	err := eventGenerator.Run()
	if err != nil {
		glog.Fatal(err)
	}
}
