package util

import (
	"flag"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"

	flags "github.com/jessevdk/go-flags"
)

// ParseArgs needs a struct compatible to jeddevdk/go-flags and will fill it
// based on CLI parameters.
func ParseArgs(options interface{}) {
	_, err := flags.ParseArgs(options, os.Args)
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	fixGlog(options)
}

// InstallSignalHandler sends a message on sigint or sigterm
func InstallSignalHandler(stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		stop <- struct{}{}
	}()
}

// configure glog, not used for flag parsing
func fixGlog(options interface{}) {
	flag.Set("logtostderr", "true")

	verbose := reflect.ValueOf(options).Elem().FieldByName("Verbose")
	if verbose.IsValid() {
		flag.Set("v", strconv.Itoa(verbose.Interface().(int)))
	}
	flag.CommandLine.Parse([]string{})
}
