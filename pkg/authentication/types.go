package authentication

import (
	"errors"
	"net/url"
)

type OIDCConfig struct {
	IssuerUrl    *string
	ClientId     *string
	ClientSecret *string
	Audience     *string
	Enabled      *bool
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
