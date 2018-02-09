package forwarder

import (
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
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://dns.google.com/resolve", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.URL.RawQuery = r.URL.RawQuery
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
	w.Write(body)
}
