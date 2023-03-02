package authtoken

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"observability-remote-write-proxy/api"
	"path"
	"time"
)

const (
	httpAuthorizationHeader = "Authorization"
)

type TokenVerifier struct {
	client *http.Client
	config api.TokenVerificationConfig
}

func NewTokenVerifier(client *http.Client, config api.TokenVerificationConfig) TokenVerifier {
	return TokenVerifier{
		client: client,
		config: config,
	}
}

func NewDefaultTokenVerifier(config api.TokenVerificationConfig) (TokenVerifier, error) {
	client, err := defaultHTTPClient()
	if err != nil {
		return TokenVerifier{}, err
	}
	return NewTokenVerifier(&client, config), nil
}

func (t *TokenVerifier) Enabled() bool {
	return t.config.Enabled != nil && *t.config.Enabled
}

func (t *TokenVerifier) GetAuthenticationToken(r *http.Request) string {
	if val, ok := r.Header[httpAuthorizationHeader]; ok {
		return val[0]
	}
	return ""
}

func (t *TokenVerifier) ValidateToken(url *url.URL, clusterID string, token string) error {
	req, err := http.NewRequest(http.MethodPost, url.String(), nil)
	if err != nil {
		return err
	}

	req.URL.Path = path.Join(req.URL.Path, clusterID)

	req.Header.Add(httpAuthorizationHeader, token)
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected status code from token verification, got %v", resp.StatusCode))
	}

	return nil
}

func defaultTLSConfig() (tls.Config, error) {
	// We get the system's certificate pool
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return tls.Config{}, nil
	}

	tlsConfig := tls.Config{
		RootCAs: rootCAs,
	}

	return tlsConfig, nil
}

func defaultTransport() (http.RoundTripper, error) {
	// our default transport is a clone of http.DefaultTranport to
	// which we add some customizations
	transport := http.DefaultTransport.(*http.Transport).Clone()

	tlsConfig, err := defaultTLSConfig()
	if err != nil {
		return nil, err
	}

	transport.TLSClientConfig = &tlsConfig

	return transport, nil
}

func defaultHTTPClient() (http.Client, error) {
	transport, err := defaultTransport()
	if err != nil {
		return http.Client{}, err
	}

	httpClient := http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	return httpClient, nil
}
