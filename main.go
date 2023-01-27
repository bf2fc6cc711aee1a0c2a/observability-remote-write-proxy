package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/pkg/remotewrite"
)

const (
	prefixHeader = "X-Forwarded-Prefix"
)

func main() {
	upstreamUrl, err := url.Parse("http://localhost:9091/api/v1/write")
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
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Println(err)
		},
	}

	proxy.Transport = http.DefaultTransport
	http.Handle("/", http.HandlerFunc(handleRequestSuccess))
	http.Handle("/", &proxy)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func handleRequestSuccess(w http.ResponseWriter, r *http.Request) {
	log.Println("request received")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeRequest, err := remotewrite.DecodeWriteRequest(r.Body)
	if err != nil {
		log.Printf("error receiving remote write request: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	clusterIDs := remotewrite.FindClusterIDs(writeRequest)
	if len(clusterIDs) > 1 {
		log.Printf("request contains multiple cluster IDs: %v", clusterIDs)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("there were %v time series in the request", len(writeRequest.Timeseries))
	w.WriteHeader(http.StatusNoContent)
	return
}
