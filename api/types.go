package api

import (
	"errors"
	"net/url"
)

type ProxyConfig struct {
	ProxyPort   *int
	MetricsPort *int
	ForwardUrl  *string
}

type OIDCConfig struct {
	IssuerUrl    *string
	ClientId     *string
	ClientSecret *string
	Audience     *string
	Enabled      *bool
}

type TokenVerificationConfig struct {
	Enabled *bool
	Url     *string
}

func (c *OIDCConfig) Validate() {
	if !*c.Enabled {
		return
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
