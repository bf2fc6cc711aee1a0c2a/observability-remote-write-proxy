package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/coreos/go-oidc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "golang.org/x/oauth2/clientcredentials"
	"log"
	"net/http"
	"net/url"
	"observability-remote-write-proxy/api"
	"observability-remote-write-proxy/pkg/authtoken"
	"observability-remote-write-proxy/pkg/metrics"
	"observability-remote-write-proxy/pkg/proxy"
	"observability-remote-write-proxy/pkg/remotewrite"
	"os"
	"os/signal"
	"syscall"
)

var (
	proxyConfig             api.ProxyConfig
	oidcConfig              api.OIDCConfig
	tokenVerificationConfig api.TokenVerificationConfig
)

func main() {
	flag.Parse()
	oidcConfig.Validate()

	upstreamUrl, err := url.Parse(*proxyConfig.ForwardUrl)
	if err != nil {
		panic(err)
	}

	proxy, err := proxy.CreateProxy(upstreamUrl, &oidcConfig)
	if err != nil {
		panic(err)
	}

	var parsedTokenVerificationUrl *url.URL
	if *tokenVerificationConfig.Enabled {
		parsedTokenVerificationUrl, err = url.Parse(*tokenVerificationConfig.Url)
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("auth token verification is disabled")
	}

	metricsServer := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", *proxyConfig.MetricsPort),
		Handler: promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}),
	}

	proxyServer := http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%v", *proxyConfig.ProxyPort),
		Handler: struct {
			http.HandlerFunc
		}{
			func(w http.ResponseWriter, r *http.Request) {
				// GET /healthcheck can be used as readiness probe
				if r.Method == http.MethodGet && r.URL.Path == "/healthcheck" {
					w.WriteHeader(http.StatusOK)
					return
				}

				// POST / is the only accepted endpoint for incoming write requests
				if r.Method != http.MethodPost || r.URL.Path != "/" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				metrics.IncomingRequestCount.WithLabelValues().Inc()

				// extract the remote write request from the http request
				remoteWriteRequest, err := remotewrite.DecodeWriteRequest(r)
				if err != nil {
					log.Printf("error decoding remote write request: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// validate the cluster ids contained in the remote write request
				clusterId, err := remotewrite.ValidateRequest(remoteWriteRequest)
				if err != nil {
					log.Printf("error validating remote write request: %v", err)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				if *tokenVerificationConfig.Enabled {
					token := authtoken.GetAuthenticationToken(r)
					if token != "" {
						err = authtoken.ValidateToken(parsedTokenVerificationUrl, clusterId, token)
						if err != nil {
							log.Printf("error validating auth token: %v", err)
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
					} else {
						log.Println("missing auth token in incoming request")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}

				// copy the remote write request back onto the http request
				err = remotewrite.PopulateRequestBody(remoteWriteRequest, r)
				if err != nil {
					log.Printf("error copying remote write request: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// after successful validation, forward the request
				proxy.ServeHTTP(w, r)
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE, syscall.SIGABRT)
	defer stop()

	// metrics server
	log.Println(fmt.Sprintf("metrics server listening on 0.0.0.0:%v", *proxyConfig.MetricsPort))
	go metricsServer.ListenAndServe()
	defer metricsServer.Shutdown(ctx)

	// proxy server
	log.Println(fmt.Sprintf("proxy server listening on 0.0.0.0:%v", *proxyConfig.ProxyPort))
	go proxyServer.ListenAndServe()
	defer proxyServer.Shutdown(ctx)
	<-ctx.Done()
}

func init() {
	proxyConfig.ProxyPort = flag.Int("proxy.listen.port", 8080, "port on which the proxy listens for incoming requests")
	proxyConfig.MetricsPort = flag.Int("proxy.metrics.port", 9090, "port on which proxies own metrics are exposed")
	proxyConfig.ForwardUrl = flag.String("proxy.forwardUrl", "", "url to forward requests to after successful validation")
	oidcConfig.IssuerUrl = flag.String("oidc.issuerUrl", "", "url of the token issuer for outgoing requests")
	oidcConfig.ClientId = flag.String("oidc.clientId", "", "client ID used to fetch tokens for outgoing requests")
	oidcConfig.ClientSecret = flag.String("oidc.clientSecret", "", "client secret used to fetch tokens for outgoing requests")
	oidcConfig.Audience = flag.String("oidc.audience", "", "oid audience sent to the token issuer")
	oidcConfig.Enabled = flag.Bool("oidc.enabled", false, "enable token authentication for outgoing requests")
	tokenVerificationConfig.Url = flag.String("token.verification.url", "", "url to validate cluster IDs and tokens of incoming requests")
	tokenVerificationConfig.Enabled = flag.Bool("token.verification.enabled", false, "enable validating cluster IDs and tokens of incoming requests")
}
