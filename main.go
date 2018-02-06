package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

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

	var (
		rr  dns.RR
		rrh dns.RR_Header
	)

	for _, q := range r.Question {
		domain := q.Name
		qtype := q.Qtype

		hdr, err := resolveByHttp(domain, qtype)
		if err != nil {
			dns.HandleFailed(w, r)
			return
		}
		msg.Authoritative = hdr.TC

		for _, a := range hdr.Answer {
			rrh = dns.RR_Header{
				Name:   a["name"].(string),
				Rrtype: qtype,
				Class:  dns.ClassINET,
				Ttl:    uint32(a["TTL"].(float64)),
			}

			switch qtype {
			case dns.TypeA:
				rr = &dns.A{
					Hdr: rrh,
					A:   net.ParseIP(a["data"].(string)),
				}
			case dns.TypeCNAME:
				rr = &dns.CNAME{
					Hdr:    rrh,
					Target: a["data"].(string),
				}
			case dns.TypeAAAA:
				rr = &dns.AAAA{
					Hdr:  rrh,
					AAAA: net.ParseIP(a["data"].(string)),
				}
			case dns.TypeANY:
				rr = &dns.ANY{rrh}
			case dns.TypeMX:
				d, _ := a["data"].(string)
				ds := strings.Split(d, " ")
				prf, _ := strconv.Atoi(ds[0])

				rr = &dns.MX{
					Hdr:        rrh,
					Preference: uint16(prf),
					Mx:         ds[1],
				}
			}

			msg.Answer = append(rsg.Answer, rr)
		}
	}

	w.WriteMsg(&msg)
}

func resolveByHttp(name string, qtype uint16) (*HttpDnsResponse, error) {
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
	q.Add("type", strconv.Itoa(int(qtype)))
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
