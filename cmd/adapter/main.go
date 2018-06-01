package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/miekg/dns"
	"github.com/tranch/httpdns/adapter"
)

var appVersion string

func main() {
	var (
		api     = flag.String("apiUrl", "https://dns.google.com/resolve", "The API URL of HTTP DNS server")
		host    = flag.String("host", "127.0.0.1", "host")
		port    = flag.Int("port", 5353, "port")
		encode  = flag.Bool("encode", false, "encode request with base64")
		version = flag.Bool("v", false, "version")
	)

	flag.Parse()

	if *version {
		fmt.Println("Version:", appVersion)
		return
	}

	srv := &dns.Server{Addr: fmt.Sprintf("%s:%d", *host, *port), Net: "udp"}
	srv.Handler = &adapter.Handle{API: *api, Encode: *encode}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
