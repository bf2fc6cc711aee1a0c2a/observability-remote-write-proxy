package proxy

import (
	"context"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/http/httputil"
	"net/url"
	"observability-remote-write-proxy/api"
	"observability-remote-write-proxy/pkg/metrics"
)

const (
	prefixHeader = "X-Forwarded-Prefix"
)

type instrumentedRoundTripper struct {
	base http.RoundTripper
}

func (i instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	metrics.OutgoingRequestCount.WithLabelValues().Inc()
	return i.base.RoundTrip(r)
}

func CreateProxy(upstreamUrl *url.URL, oidcConfig *api.OIDCConfig) (*httputil.ReverseProxy, error) {
	proxy := &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = upstreamUrl.Scheme
			request.Host = upstreamUrl.Host
			request.URL.Host = upstreamUrl.Host
			request.URL.Path = upstreamUrl.Path
			request.Header.Add(prefixHeader, "/")
		},
	}

	rt := instrumentedRoundTripper{
		base: http.DefaultTransport,
	}

	if *oidcConfig.Enabled {
		provider, err := oidc.NewProvider(context.Background(), *oidcConfig.IssuerUrl)
		if err != nil {
			return nil, err
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
			Base:   rt,
		}
	} else {
		proxy.Transport = rt
	}

	return proxy, nil
}
