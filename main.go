package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/pkg/remotewrite"
)

const (
	prefixHeader = "X-Forwarded-Prefix"
)

var (
	Upstream = "http://localhost:9091/api/v1/write"
	address  *string
	upstream *string
)

func main() {
	upstreamUrl, err := url.Parse(*upstream)
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
		err := remotewrite.ValidateRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		return
	}
}

func init() {
	address = flag.String("address", ":8080", "port to listen on")
	upstream = flag.String("upstream", Upstream, "upstream to proxy to")
}
