package forwarder

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

	var (
		fName     string
		fType     string
		isEncoded bool
	)

	switch r.Method {
	case "GET":
		q := r.URL.Query() // Request query
		fName = q.Get("name")
		fType = q.Get("type")
		isEncoded = q.Get("encoded") == "yes"
	case "POST":
		fName = r.FormValue("name")
		fType = r.FormValue("type")
		isEncoded = r.FormValue("encoded") == "yes"
	default:
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	if isEncoded {
		domain, err := base64.StdEncoding.DecodeString(fName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fName = string(domain)
	}

	req, err := http.NewRequest("GET", "https://dns.google.com/resolve", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fq := req.URL.Query() // Forward query
	fq.Add("name", fName)
	fq.Add("type", fType)
	req.URL.RawQuery = fq.Encode()

	client := &http.Client{}
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

	for fwName, fwVals := range resp.Header {
		for _, fwVal := range fwVals {
			w.Header().Set(fwName, fwVal)
		}
	}

	if isEncoded {
		w.Header().Set("Content-Type", "text/plain")
		bodyEncoded := base64.StdEncoding.EncodeToString(body)
		response = []byte(bodyEncoded)
	} else {
		response = body
	}

	w.Write(response)
}
