package api

import (
	"testing"
)

var (
	OIDCNotEnabled     = false
	OIDCIsEnabled      = true
	emptyFIlename      = ""
	OIDCConfigFilename = "./secrets/oidc-config.yaml.sample"
)

func TestOIDCConfig_Validate(t *testing.T) {
	type fields struct {
		Enabled    *bool
		Filename   *string
		Attributes OIDCAttributes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "should return nil if oidc is not enabled",
			fields: fields{
				Enabled: &OIDCNotEnabled,
			},
			wantErr: false,
		},
		{
			name: "should return error if oidc is enabled but no filename is provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &emptyFIlename,
			},
			wantErr: true,
		},
		{
			name: "should return nil if oidc is enabled and filename is provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
			},
			wantErr: true,
		},
		{
			name: "should return nil if oidc is enabled and filename is provided, but file does not exist",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
			},
			wantErr: true,
		},
		{
			name: "should return an error if oidc is enabled but the issuer url is not provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
				Attributes: OIDCAttributes{
					IssuerURL: "",
				},
			},
			wantErr: true,
		},
		{
			name: "should return an error if oidc is enabled but the client id is not provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
				Attributes: OIDCAttributes{
					IssuerURL: "https://test.com",
					ClientID:  "",
				},
			},
			wantErr: true,
		},
		{
			name: "should return an error if oidc is enabled but the client secret is not provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
				Attributes: OIDCAttributes{
					IssuerURL:    "https://test.com",
					ClientID:     "test",
					ClientSecret: "",
				},
			},
			wantErr: true,
		},
		{
			name: "should return nil if oidc is enabled and all required fields are provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
				Attributes: OIDCAttributes{
					IssuerURL:    "https://test.com",
					ClientID:     "test",
					ClientSecret: "test",
				},
			},
			wantErr: true,
		},
		{
			name: "should return nil if oidc is enabled and all required fields are provided along with optional fields",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &OIDCConfigFilename,
				Attributes: OIDCAttributes{
					IssuerURL:    "https://test.com",
					ClientID:     "test",
					ClientSecret: "test",
					Audience:     "test",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &OIDCConfig{
				Enabled:    tt.fields.Enabled,
				Filename:   tt.fields.Filename,
				Attributes: tt.fields.Attributes,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("OIDCConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
