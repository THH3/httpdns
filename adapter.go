package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/miekg/dns"
)

const HttpDnsAPI = "https://dns.google.com/resolve"

type HttpDnsResponse struct {
	Status   int
	TC       bool
	RD       bool
	RA       bool
	AD       bool
	CD       bool
	Question []map[string]interface{}
	Answer   []map[string]interface{}
}

type Handle struct{}

func (h *Handle) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	for _, q := range r.Question {
		domain := q.Name
		rtype := q.Qtype

		hdr, err := resolveByHttp(domain, rtype)
		if err != nil || hdr.Status != 0 {
			dns.HandleFailed(w, r)
			return
		}

		msg.Authoritative = hdr.TC

		for _, a := range hdr.Answer {
			v := dns.Fqdn(domain) + "\t"
			v += strconv.Itoa(int(a["TTL"].(float64))) + "\t"
			v += dns.ClassToString[dns.ClassINET] + "\t"
			v += dns.TypeToString[uint16(a["type"].(float64))] + "\t"
			v += a["data"].(string)

			if rr, err := dns.NewRR(v); err == nil {
				msg.Answer = append(msg.Answer, rr)
			}
		}
	}

	w.WriteMsg(&msg)
}

func resolveByHttp(name string, rtype uint16) (*HttpDnsResponse, error) {
	var (
		hdr *HttpDnsResponse
		err error
	)

	client := &http.Client{}

	req, err := http.NewRequest("GET", HttpDnsAPI, nil)
	if err != nil {
		log.Println("Unable to make request:", err.Error())
		return hdr, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("type", strconv.Itoa(int(rtype)))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println("HTTP DNS API Error:", err.Error())
		return hdr, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body:", err.Error())
		return hdr, err
	}

	err = json.Unmarshal(body, &hdr)
	if err != nil {
		log.Println("Unable to parse response as JSON:", err.Error())
		return hdr, err
	}

	return hdr, nil
}

func main() {
	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 5353, "port")
	flag.Parse()

	bindAddr := *host + ":" + strconv.Itoa(*port)

	srv := &dns.Server{Addr: bindAddr, Net: "udp"}
	srv.Handler = &Handle{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
