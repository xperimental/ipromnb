package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/xperimental/ipromnb/kernel"
	"github.com/xperimental/ipromnb/scaffold"
)

var (
	configFile string
	serverURL  string
)

func main() {
	flag.StringVar(&configFile, "connection-file", "", "Path to connection file.")
	flag.StringVar(&serverURL, "server-url", "", "Default Prometheus server.")
	flag.Parse()

	if configFile == "" {
		glog.Fatal("Need to provide a connection file.")
	}

	kernel := kernel.New(serverURL)

	server, err := scaffold.NewServer(configFile, kernel)
	if err != nil {
		glog.Fatalf("Error creating server: %s", err)
	}

	server.Loop()
}
