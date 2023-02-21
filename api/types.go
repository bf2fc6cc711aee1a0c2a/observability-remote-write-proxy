package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
)

type ProxyConfig struct {
	ProxyPort   *int
	MetricsPort *int
	ForwardUrl  *string
}

type OIDCConfig struct {
	IssuerUrl    *string `json:"issuer_url"`
	ClientId     *string `json:"client_id"`
	ClientSecret *string `json:"client_secret"`
	Audience     *string `json:"audience"`
	Enabled      *bool
	Filename     *string
}

type TokenVerificationConfig struct {
	Enabled *bool
	Url     *string
}

func (c *OIDCConfig) ReadAndValidate() {
	if !*c.Enabled {
		return
	} else {
		configFile, err := os.Open(*c.Filename)
		if err != nil {
			panic(err)
		}
		data, err := io.ReadAll(configFile)
		if err == nil && data != nil {
			err = json.Unmarshal(data, &*c)
		}
	}

	if *c.IssuerUrl == "" {
		panic(errors.New("token issuer url required"))
	}

	_, err := url.Parse(*c.IssuerUrl)
	if err != nil {
		panic(err)
	}

	if *c.ClientSecret == "" || *c.ClientId == "" {
		panic(errors.New("client id and secret are required"))
	}
}
