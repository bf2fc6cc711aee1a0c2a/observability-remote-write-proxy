package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/pkg/remotewrite"
	"os"
)

const (
	prefixHeader = "X-Forwarded-Prefix"
)

var (
	PORT     = os.Getenv("PORT")
	upstream = os.Getenv("UPSTREAM")
)

func main() {
	upstreamUrl, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}

	proxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = upstreamUrl.Scheme
			request.Host = upstreamUrl.Host
			request.URL.Host = upstreamUrl.Host
			request.URL.Path = upstreamUrl.Path
			request.Header.Add(prefixHeader, "/")
		},
	}

	proxy.Transport = http.DefaultTransport
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := validateRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})
	err = http.ListenAndServe(PORT, nil)
	if err != nil {
		return
	}
}

func validateRequest(r *http.Request) error {
	log.Println("request received")

	if r.Method != http.MethodPost {
		return errors.New("not found")
	}

	writeRequest, err := remotewrite.DecodeWriteRequest(r.Body)
	if err != nil {
		log.Printf("error receiving remote write request: %v", err.Error())
		return errors.New("error receiving remote write request")
	}

	log.Printf("there are %v time series in the request", len(writeRequest.Timeseries))

	clusterIDs := remotewrite.FindClusterIDs(writeRequest)
	if len(clusterIDs) > 1 {
		log.Printf("request contains multiple cluster IDs: %v", clusterIDs)
		return errors.New("request contains multiple cluster IDs")
	}

	return nil
}
