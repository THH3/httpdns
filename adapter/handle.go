package adapter

import (
	"context"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

// special a upstream DNS server to resolve API server's domain
var UpstreamServerAddr string

// wait seconds for API request
var RequestTimeout = 5 * time.Second

type apiResponse struct {
	Status   int  // Standard DNS response code (32 bit integer).
	TC       bool // Whether the response is truncated
	RD       bool // Always true for Google Public DNS
	RA       bool // Always true for Google Public DNS
	AD       bool // Whether all response data was validated with DNSSEC
	CD       bool // Whether the client asked to disable DNSSEC
	Question []map[string]interface{}
	Answer   []apiResponseAnswer
}

type apiResponseAnswer struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
	TTL  int
	Data string `json:"data"`
}

func (a *apiResponseAnswer) String() string {
	v := dns.Fqdn(a.Name) + "\t"
	v += strconv.Itoa(a.TTL) + "\t"
	v += dns.ClassToString[dns.ClassINET] + "\t"
	v += dns.TypeToString[a.Type] + "\t"
	v += a.Data

	return v
}

type Handle struct {
	API    string
	Encode bool
}

func (h *Handle) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	for _, q := range r.Question {
		hdr, err := h.ResolveByHttp(q.Name, q.Qtype)
		if err != nil || hdr.Status != 0 {
			dns.HandleFailed(w, r)
			return
		}

		for _, a := range hdr.Answer {
			if rr, err := dns.NewRR(a.String()); err == nil {
				msg.Answer = append(msg.Answer, rr)
			}
		}
	}

	w.WriteMsg(&msg)
}

func (h *Handle) ResolveByHttp(name string, rtype uint16) (*apiResponse, error) {
	var (
		hdr *apiResponse
		err error
	)

	// Must provide a DNS server to resolve the API server domain's IP address
	// when running as a default DNS server.
	http.DefaultTransport.(*http.Transport).DialContext = dialContext

	client := &http.Client{Timeout: RequestTimeout}

	req, err := http.NewRequest("GET", h.API, nil)
	if err != nil {
		log.Println("Unable to make request:", err)
		return hdr, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("type", strconv.Itoa(int(rtype)))

	if h.Encode {
		q.Add("encoded", "yes")
		q.Set("name", base64.StdEncoding.EncodeToString([]byte(name)))
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Unable to request API server:", err)
		return hdr, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body:", err)
		return hdr, err
	}

	var data []byte

	if h.Encode {
		data, err = base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			log.Println("Unable decode response from base64:", err)
			return hdr, err
		}
	} else {
		data = body
	}

	if err = json.Unmarshal([]byte(data), &hdr); err != nil {
		log.Println("Unable to parse response as JSON:", err)
		return hdr, err
	}

	return hdr, nil
}

func dnsDialer(ctx context.Context, network, address string) (net.Conn, error) {
	d := net.Dialer{}

	if UpstreamServerAddr == "" {
		UpstreamServerAddr = address
	}

	return d.DialContext(ctx, network, UpstreamServerAddr)
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial:     dnsDialer,
		},
	}

	return d.DialContext(ctx, network, addr)
}
