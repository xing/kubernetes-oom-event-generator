package util

import (
	"flag"
	"os"
	"testing"
)

func TestParseArgs(t *testing.T) {
	os.Args = []string{"bin/test", "--host", "kubernetes.io"}

	var options struct {
		Host string `long:"host" required:"true"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	if options.Host != "kubernetes.io" {
		t.Errorf("Host was %s", options.Host)
	}
}

func TestParseArgsWithVerboseButNotSet(t *testing.T) {
	os.Args = []string{"bin/test"}

	var options struct {
		Verbose int `short:"v"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	if options.Verbose != 0 {
		t.Errorf("Verbose was %d", options.Verbose)
	}
}

func TestParseArgsWithVerboseSet(t *testing.T) {
	os.Args = []string{"bin/test", "-v", "3"}

	var options struct {
		Verbose int `short:"v"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	if options.Verbose != 3 {
		t.Errorf("Verbose was %d", options.Verbose)
	}
}
