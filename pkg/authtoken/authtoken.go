package authtoken

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	httpAuthorizationHeader = "Authorization"
)

// CACertificate represents the content
// of a CA certificate
type CAPEMCertificateRawBytes []byte

func (c *CAPEMCertificateRawBytes) ToByteSlice() []byte {
	return []byte(*c)
}

func ToCAPEMCertificateRawBytes(certRawBytes []byte) CAPEMCertificateRawBytes {
	return CAPEMCertificateRawBytes(certRawBytes)
}

type TokenVerifierOptions struct {
	Enabled       bool
	URL           url.URL
	CACertEnabled bool
	CACertRaw     CAPEMCertificateRawBytes
}

type TokenVerifier struct {
	client  http.Client
	options TokenVerifierOptions
}

func NewDefaultTokenVerifier(options TokenVerifierOptions) (TokenVerifier, error) {
	client, err := defaultHTTPClient(options)
	if err != nil {
		return TokenVerifier{}, err
	}
	return newTokenVerifier(client, options), nil
}

func newTokenVerifier(client http.Client, options TokenVerifierOptions) TokenVerifier {
	return TokenVerifier{
		client:  client,
		options: options,
	}
}

func (t *TokenVerifier) Enabled() bool {
	return t.options.Enabled
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

func defaultTLSConfig(options TokenVerifierOptions) (tls.Config, error) {
	var tlsConfig tls.Config
	if !options.CACertEnabled {
		return tlsConfig, nil
	}

	// We get the system's certificate pool
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return tls.Config{}, nil
	}

	// Append the CA cert to the system's certificate pool
	if ok := rootCAs.AppendCertsFromPEM(options.CACertRaw); !ok {
		return tls.Config{}, fmt.Errorf("failed to parse token verifier CA Certificate as a PEM encoded certificate")
	}

	tlsConfig.RootCAs = rootCAs

	return tlsConfig, nil
}

func defaultTransport(options TokenVerifierOptions) (http.RoundTripper, error) {
	// our default transport is a clone of http.DefaultTranport to
	// which we add some customizations
	transport := http.DefaultTransport.(*http.Transport).Clone()

	tlsConfig, err := defaultTLSConfig(options)
	if err != nil {
		return nil, err
	}

	transport.TLSClientConfig = &tlsConfig

	return transport, nil
}

func defaultHTTPClient(options TokenVerifierOptions) (http.Client, error) {
	transport, err := defaultTransport(options)
	if err != nil {
		return http.Client{}, err
	}

	httpClient := http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	return httpClient, nil
}
