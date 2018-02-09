package main

import (
	"flag"

	"github.com/tr4nch/httpdns/forwarder"
)

var appVersion string

func main() {
	var (
		host    = flag.String("host", "localhost", "host to binding")
		port    = flag.Int("port", 8453, "port to listen")
		version = flag.Bool("v", false, "version")
	)

	flag.Parse()

	if *version {
		println("Version:", appVersion)
		return
	}

	srv := &forwarder.Server{*host, *port}
	srv.Run()
}
