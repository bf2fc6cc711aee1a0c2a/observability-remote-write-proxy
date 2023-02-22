package main

import (
	"context"
	"flag"
	"fmt"
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

	_ "github.com/coreos/go-oidc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "golang.org/x/oauth2/clientcredentials"
)

var (
	proxyConfig             api.ProxyConfig
	oidcConfig              api.OIDCConfig
	tokenVerificationConfig api.TokenVerificationConfig
)

func main() {
	flag.Parse()
	err := oidcConfig.ReadAndValidate()
	if err != nil {
		log.Printf("Error reading and validating OIDC config: %v", err)
		os.Exit(1)
	}

	upstreamUrl, err := url.Parse(*proxyConfig.ForwardUrl)
	if err != nil {
		log.Printf("Error parsing upstream url: %v", err)
		os.Exit(1)
	}

	proxy, err := proxy.CreateProxy(upstreamUrl, &oidcConfig)
	if err != nil {
		log.Printf("Error creating proxy: %v", err)
		os.Exit(1)
	}

	var parsedTokenVerificationUrl *url.URL
	if *tokenVerificationConfig.Enabled {
		parsedTokenVerificationUrl, err = url.Parse(*tokenVerificationConfig.Url)
		if err != nil {
			log.Printf("Error parsing token verification url: %v", err)
			os.Exit(1)
		}
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
					log.Printf("error validating the remote write request: %v", err)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				if *tokenVerificationConfig.Enabled {
					token := authtoken.GetAuthenticationToken(r)
					if token != "" {
						err = authtoken.ValidateToken(parsedTokenVerificationUrl, clusterId, token)
						if err != nil {
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
					} else {
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
	proxyConfig.MetricsPort = flag.Int("proxy.metrics.port", 9090, "port on which proxy metrics are exposed")
	proxyConfig.ForwardUrl = flag.String("proxy.forwardUrl", "", "url to forward requests to")
	oidcConfig.Enabled = flag.Bool("oidc.enabled", false, "enable oidc authentication")
	oidcConfig.Filename = flag.String("oidc.filename", "", "path to oidc configuration file")
	tokenVerificationConfig.Url = flag.String("token.verification.url", "", "url to validate data plane tokens")
	tokenVerificationConfig.Enabled = flag.Bool("token.verification.enabled", false, "enable data plane token verification")
}
