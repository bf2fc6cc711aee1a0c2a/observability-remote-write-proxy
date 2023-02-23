package api

import (
	"errors"
	"gopkg.in/yaml.v3"
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
	Enabled    *bool
	Filename   *string
	Attributes OIDCAttributes
}

type OIDCAttributes struct {
	IssuerURL    string `yaml:"issuer_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Audience     string `yaml:"audience"`
}

type TokenVerificationConfig struct {
	Enabled *bool
	Url     *string
}

func (c *OIDCConfig) Read() error {
	configFile, err := os.Open(*c.Filename)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &c.Attributes)
	if err != nil {
		return err
	}

	return nil
}

func (c *OIDCConfig) Validate() error {
	if !*c.Enabled {
		return nil
	}

	err := c.Read()
	if err != nil {
		return err
	}

	if c.Attributes.IssuerURL == "" {
		return err
	}

	_, err = url.Parse(c.Attributes.IssuerURL)
	if err != nil {
		return err
	}

	if c.Attributes.ClientSecret == "" || c.Attributes.ClientID == "" {
		return errors.New("client id and secret are required")
	}
	return nil
}
