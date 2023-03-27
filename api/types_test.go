package api

import (
	"os"
	"testing"
)

func TestOIDCConfig_Validate(t *testing.T) {
	OIDCNotEnabled := false
	OIDCIsEnabled := true
	emptyFilename := ""
	filenameThatDoesntExist := "doesnt-exist.yaml"
	fullConfigFile, _ := os.CreateTemp("", "")
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(fullConfigFile.Name())
	filenameComplete := fullConfigFile.Name()
	yamlDataComplete := `
issuer_url: https://example.com
client_id: client-id
client_secret: client-secret
`
	_, _ = fullConfigFile.Write([]byte(yamlDataComplete))
	err := fullConfigFile.Close()
	if err != nil {
		return
	}
	emptyConfigFile, _ := os.CreateTemp("", "")
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(emptyConfigFile.Name())
	filenameNotComplete := emptyConfigFile.Name()
	yamlDataNotComplete := ``
	_, _ = emptyConfigFile.Write([]byte(yamlDataNotComplete))
	err = emptyConfigFile.Close()
	if err != nil {
		return
	}
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
				Filename: &emptyFilename,
			},
			wantErr: true,
		},
		{
			name: "should return err if oidc is enabled and filename is provided, but file does not exist",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &filenameThatDoesntExist,
			},
			wantErr: true,
		},
		{
			name: "should return an error if oidc is enabled but the client id is not provided",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &filenameNotComplete,
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
				Filename: &filenameNotComplete,
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
				Filename: &filenameComplete,
				Attributes: OIDCAttributes{
					IssuerURL:    "https://test.com",
					ClientID:     "test",
					ClientSecret: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "should return nil if oidc is enabled and all required fields are provided along with optional fields",
			fields: fields{
				Enabled:  &OIDCIsEnabled,
				Filename: &filenameComplete,
				Attributes: OIDCAttributes{
					IssuerURL:    "https://test.com",
					ClientID:     "test",
					ClientSecret: "test",
					Audience:     "test",
				},
			},
			wantErr: false,
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
