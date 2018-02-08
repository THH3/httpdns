package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

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

func main() {
	host := flag.String("host", "localhost", "host to binding")
	port := flag.Int("port", 8453, "port to listen")
	flag.Parse()

	bindAddr := *host + ":" + strconv.Itoa(*port)
	log.Println("Listening on", bindAddr)

	http.HandleFunc("/resolve", resolveHandle)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
