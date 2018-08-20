package main

import (
	"fmt"
	"log"
	"github.com/miekg/dns"
	"github.com/tranch/httpdns/adapter"
	"github.com/jessevdk/go-flags"
)

var appVersion string

type options struct {
	Api       string `short:"a" long:"api" description:"API URL"`
	LocalAddr string `short:"b" description:"Special local address to bind"`
	Encode    bool   `short:"e" long:"enc" description:"Encoding request body with base64"`
	Upstream  string `short:"u" long:"upstream" description:"Special upstream DNS server to resolve API server's domain"`
	Version   bool   `short:"v" long:"version" description:"Show version"`
}

func main() {
	opts := &options{
		Api:       "https://dns.google.com/resolve",
		LocalAddr: ":53",
		Encode:    false,
		Version:   false,
	}

	_, err := flags.Parse(opts)
	if err != nil {
		return
	}

	if opts.Version {
		fmt.Println("Version:", appVersion)
		return
	}

	if opts.Upstream != "" {
		adapter.UpstreamServerAddr = opts.Upstream
	}

	srv := &dns.Server{Addr: opts.LocalAddr, Net: "udp"}
	srv.Handler = &adapter.Handle{API: opts.Api, Encode: opts.Encode}

	log.Printf("Listening on %s via %s %s base64 encoding\n", opts.LocalAddr, opts.Api,
		map[bool]string{true: "with", false: "without"}[opts.Encode])

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
