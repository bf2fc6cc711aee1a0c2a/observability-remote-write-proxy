package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/pkg/remotewrite"
	"strings"
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
			// Set the Host at both request and request.URL objects.
			request.Host = upstreamUrl.Host
			request.URL.Host = upstreamUrl.Host
			// Derive path from the paths of configured URL and request URL.
			request.URL.Path = upstreamUrl.Path
			// Add prefix header with value "/", since from a client's perspective
			// we are forwarding /<anything> to /<cfg.upstream.url.Path>/<anything>.
			request.Header.Add(prefixHeader, "/")
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Println(err)
		},
	}

	proxy.Transport = http.DefaultTransport

	http.Handle("/", &proxy)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	apath := a.EscapedPath()
	bpath := b.EscapedPath()
	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")
	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
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
