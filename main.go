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
	err := initializeTokenVerificationConfig(tokenVerificationConfig)
	if err != nil {
		log.Printf("error initializing token verification config: %s", err)
		os.Exit(1)
	}

	err = oidcConfig.Validate()
	if err != nil {
		log.Printf("error reading and validating OIDC config: %v", err)
		os.Exit(1)
	}

	upstreamUrl, err := url.Parse(*proxyConfig.ForwardUrl)
	if err != nil {
		log.Printf("error parsing upstream url: %v", err)
		os.Exit(1)
	}

	proxy, err := proxy.CreateProxy(upstreamUrl, &oidcConfig)
	if err != nil {
		log.Printf("error creating proxy: %v", err)
		os.Exit(1)
	}

	var parsedTokenVerificationUrl *url.URL
	if *tokenVerificationConfig.Enabled {
		parsedTokenVerificationUrl, err = url.Parse(*tokenVerificationConfig.Url)
		if err != nil {
			log.Printf("error parsing token verification url: %v", err)
			os.Exit(1)
		}
	}

	metricsServer := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", *proxyConfig.MetricsPort),
		Handler: promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}),
	}

	tokenVerifierOptions, err := buildTokenVerifierOptions(tokenVerificationConfig)
	if err != nil {
		log.Printf("error building token verifier options: %v", err)
		os.Exit(1)
	}
	tokenVerifier, err := authtoken.NewDefaultTokenVerifier(tokenVerifierOptions)
	if err != nil {
		log.Printf("error creating token verifier: %v", err)
		os.Exit(1)
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
					log.Println("invalid proxy request, method must be POST and path must be /")
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

				clusterID := ""
				if !remotewrite.IsMetadataRequest(remoteWriteRequest) {
					// validate the cluster ids contained in the remote write request
					clusterID, err = remotewrite.ValidateRequest(remoteWriteRequest)
					if err != nil {
						log.Printf("error validating remote write request: %v", err)
						w.WriteHeader(http.StatusForbidden)
						return
					}

					log.Println(fmt.Sprintf("remote write request received from '%v'", clusterID))

					if tokenVerifier.Enabled() {
						token := tokenVerifier.GetAuthenticationToken(r)
						if token == "" {
							log.Println(fmt.Sprintf("auth token missing in request from '%v'", clusterID))
							w.WriteHeader(http.StatusBadRequest)
							return
						}

						err = tokenVerifier.ValidateToken(parsedTokenVerificationUrl, clusterID, token)
						if err != nil {
							log.Println(fmt.Sprintf("error validating auth token from '%v': %v", clusterID, err.Error()))
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
					}
				} else {
					log.Println("metadata request received, checks skipped")
				}

				// copy the remote write request back onto the http request
				err = remotewrite.PopulateRequestBody(remoteWriteRequest, r)
				if err != nil {
					log.Printf("error copying remote write request from '%v': %v", clusterID, err)
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

func initializeTokenVerificationConfig(config api.TokenVerificationConfig) error {
	if *config.CACertEnabled {
		reader, err := os.Open(*tokenVerificationConfig.CACertFilePath)
		if err != nil {
			return err
		}
		tokenVerificationConfig.CACertReader = reader
	}

	return nil
}

func buildTokenVerifierOptions(config api.TokenVerificationConfig) (authtoken.TokenVerifierOptions, error) {
	var options authtoken.TokenVerifierOptions
	if *config.Enabled {
		options.Enabled = true
		parsedURL, err := url.Parse(*tokenVerificationConfig.Url)
		if err != nil {
			return options, fmt.Errorf("error parsing token verification url: %v", err)
		}
		options.URL = *parsedURL
	}

	if *config.CACertEnabled {
		caCertRawBytes, err := config.ReadCACert()
		if err != nil {
			return options, fmt.Errorf("error reading token verification CA Certificate: %v", err)
		}
		options.CACertEnabled = true
		options.CACertRaw = caCertRawBytes
	}

	return options, nil
}

func init() {
	proxyConfig.ProxyPort = flag.Int("proxy.listen.port", 8080, "port on which the proxy listens for incoming requests")
	proxyConfig.MetricsPort = flag.Int("proxy.metrics.port", 9090, "port on which proxy metrics are exposed")
	proxyConfig.ForwardUrl = flag.String("proxy.forwardUrl", "", "url to forward requests to")
	oidcConfig.Enabled = flag.Bool("oidc.enabled", false, "enable oidc authentication")
	oidcConfig.Filename = flag.String("oidc.filename", "", "path to oidc configuration file")
	tokenVerificationConfig.Url = flag.String("token.verification.url", "", "url to validate data plane tokens")
	tokenVerificationConfig.Enabled = flag.Bool("token.verification.enabled", false, "enable data plane token verification")
	tokenVerificationConfig.CACertEnabled = flag.Bool("token.verification.cacert.enabled", false, "If enabled, the CA certificate provided in -token.verification.cacert.filepath will be appended to the trusted CAs store when performing token verification")
	tokenVerificationConfig.CACertFilePath = flag.String("token.verification.cacert.filename", "", "The CA certificate file path. See -token.verification.cacert.enabled for more information. If -token.verification.cacert.enabled it not set or set to false it is not used")
}
