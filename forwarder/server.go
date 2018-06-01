package forwarder

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"encoding/base64"
)

type Server struct {
	Host string
	Port int
}

func (s *Server) Run() {
	addr := s.Host + ":" + strconv.Itoa(s.Port)
	log.Println("Listening on:", addr)

	http.HandleFunc("/resolve", resolveHandle)
	err := http.ListenAndServe(addr, nil)
	log.Fatal(err)
}

func resolveHandle(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://dns.google.com/resolve", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.URL.RawQuery = r.URL.RawQuery
	q := req.URL.Query()
	encoded := q.Get("encoded")

	if encoded == "yes" {
		name := q.Get("name")
		domain, err := base64.StdEncoding.DecodeString(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q.Del("encoded")
		q.Set("name", string(domain))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []byte

	if encoded == "yes" {
		bodyEncoded := base64.StdEncoding.EncodeToString(body)
		response = []byte(bodyEncoded)
	} else {
		response = body
	}

	w.Write(response)
}
