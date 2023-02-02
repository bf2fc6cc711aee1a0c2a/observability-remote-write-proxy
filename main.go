package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/coreos/go-oidc"
	_ "github.com/coreos/go-oidc"
	"github.com/golang/snappy"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	_ "golang.org/x/oauth2/clientcredentials"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/pkg/authentication"
	"observability-remote-write-proxy/pkg/remotewrite"
)

const (
	prefixHeader = "X-Forwarded-Prefix"
)

var (
	proxyListenPort *int
	proxyForwardUrl *string
	oidcConfig      authentication.OIDCConfig
)

func populateRequestBody(rw *prometheus.WriteRequest, r *http.Request) error {
	data, err := proto.Marshal(rw)
	if err != nil {
		return err
	}
	encoded := snappy.Encode(nil, data)
	body := bytes.NewReader(encoded)
	r.Body = io.NopCloser(body)
	return nil
}

func main() {
	flag.Parse()

	oidcConfig.Validate()

	upstreamUrl, err := url.Parse(*proxyForwardUrl)
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

	if *oidcConfig.Enabled {
		provider, err := oidc.NewProvider(context.Background(), *oidcConfig.IssuerUrl)
		if err != nil {
			panic(err)
		}

		var cfg = clientcredentials.Config{
			ClientID:     *oidcConfig.ClientId,
			ClientSecret: *oidcConfig.ClientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
		}

		if *oidcConfig.Audience != "" {
			cfg.EndpointParams = map[string][]string{
				"audience": {*oidcConfig.Audience},
			}
		}

		proxy.Transport = &oauth2.Transport{
			Source: cfg.TokenSource(context.Background()),
			Base:   http.DefaultTransport,
		}
	} else {
		proxy.Transport = http.DefaultTransport
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// extract the remote write request from the http request
		remoteWriteRequest, err := remotewrite.DecodeWriteRequest(r)
		if err != nil {
			log.Printf("error decoding remote write request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("remote write request received")

		// validate the cluster ids contained in the remote write request
		err = remotewrite.ValidateRequest(remoteWriteRequest)
		if err != nil {
			log.Printf("error validating the remote write request: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// copy the remote write request back onto the http request
		err = populateRequestBody(remoteWriteRequest, r)
		if err != nil {
			log.Printf("error copying remote write request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO validate data plane auth token

		// after successful validation, forward the request
		proxy.ServeHTTP(w, r)
	})
	http.ListenAndServe(fmt.Sprintf(":%v", *proxyListenPort), nil)
}

func init() {
	proxyListenPort = flag.Int("proxy.listen.port", 8080, "port on which the proxy listens for incoming requests")
	proxyForwardUrl = flag.String("proxy.forwardUrl", "", "url to forward requests to")
	oidcConfig.IssuerUrl = flag.String("oidc.issuerUrl", "", "token issuer url")
	oidcConfig.ClientId = flag.String("oidc.clientId", "", "service account client id")
	oidcConfig.ClientSecret = flag.String("oidc.clientSecret", "", "service account client secret")
	oidcConfig.Audience = flag.String("oidc.audience", "", "oid audience")
	oidcConfig.Enabled = flag.Bool("oidc.enabled", false, "enable oidc authentication")
}
