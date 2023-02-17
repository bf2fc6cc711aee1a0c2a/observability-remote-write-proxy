package authtoken

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	headerAuthorization = "Authorization"
)

var (
	httpClient = http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Second,
	}
)

func GetAuthenticationToken(r *http.Request) string {
	if val, ok := r.Header[headerAuthorization]; ok {
		return val[0]
	}
	return ""
}

func ValidateToken(url *url.URL, clusterId string, token string) error {
	req, err := http.NewRequest(http.MethodPost, url.String(), nil)
	if err != nil {
		return err
	}

	req.URL.Path = path.Join(req.URL.Path, clusterId)

	req.Header.Add(headerAuthorization, token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected status code from token verification, got %v", resp.StatusCode))
	}

	return nil
}
