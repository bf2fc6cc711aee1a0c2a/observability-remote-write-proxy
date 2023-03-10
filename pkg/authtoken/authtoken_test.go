package authtoken

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

var (
	a       byte = 116
	b       byte = 101
	c       byte = 115
	d       byte = 116
	options      = TokenVerifierOptions{
		Enabled: true,
		URL: url.URL{
			Scheme: "https",
			Host:   "localhost:8080",
		},
		CACertEnabled: true,
		CACertRaw: CAPEMCertificateRawBytes{
			a, b, c, d,
		},
	}
	optionsWithoutCACert = TokenVerifierOptions{
		Enabled:       true,
		CACertEnabled: false,
		CACertRaw:     nil,
	}
	clientWithCACert, _ = defaultHTTPClient(options)
)

func TestCAPEMCertificateRawBytes_ToByteSlice(t *testing.T) {
	tests := []struct {
		name string
		c    *CAPEMCertificateRawBytes
		want []byte
	}{
		{
			name: "should return a byte slice",
			c: &CAPEMCertificateRawBytes{
				a, b, c, d,
			},
			want: []byte{116, 101, 115, 116},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.ToByteSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAPEMCertificateRawBytes.ToByteSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCAPEMCertificateRawBytes(t *testing.T) {
	type args struct {
		certRawBytes []byte
	}
	tests := []struct {
		name string
		args args
		want CAPEMCertificateRawBytes
	}{
		{
			name: "should return a CAPEMCertificateRawBytes",
			args: args{
				certRawBytes: []byte{116, 101, 115, 116},
			},
			want: CAPEMCertificateRawBytes{
				a, b, c, d,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCAPEMCertificateRawBytes(tt.args.certRawBytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToCAPEMCertificateRawBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDefaultTokenVerifier(t *testing.T) {
	type args struct {
		options TokenVerifierOptions
	}
	tests := []struct {
		name    string
		args    args
		want    TokenVerifier
		wantErr bool
	}{
		{
			name: "should return an error",
			args: args{
				options: options,
			},
			want:    TokenVerifier{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDefaultTokenVerifier(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefaultTokenVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultTokenVerifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newTokenVerifier(t *testing.T) {
	type args struct {
		client  http.Client
		options TokenVerifierOptions
	}
	tests := []struct {
		name string
		args args
		want TokenVerifier
	}{
		{
			name: "should return a TokenVerifier",
			args: args{
				client:  clientWithCACert,
				options: options,
			},
			want: TokenVerifier{
				client:  clientWithCACert,
				options: options,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTokenVerifier(tt.args.client, tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTokenVerifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenVerifier_Enabled(t *testing.T) {
	type fields struct {
		client  http.Client
		options TokenVerifierOptions
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "should return true",
			fields: fields{
				client:  clientWithCACert,
				options: options,
			},
			want: true,
		},
		{
			name: "should return false",
			fields: fields{
				client: clientWithCACert,
				options: TokenVerifierOptions{
					Enabled: false,
					URL: url.URL{
						Scheme: "https",
						Host:   "localhost:8080",
					},
					CACertEnabled: true,
					CACertRaw: CAPEMCertificateRawBytes{
						a, b, c, d,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TokenVerifier{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			if got := tr.Enabled(); got != tt.want {
				t.Errorf("TokenVerifier.Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenVerifier_GetAuthenticationToken(t *testing.T) {
	type fields struct {
		client  http.Client
		options TokenVerifierOptions
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "should return token",
			fields: fields{
				client:  clientWithCACert,
				options: options,
			},
			args: args{
				r: &http.Request{
					Header: http.Header{
						"Authorization": []string{"Bearer token"},
					},
				},
			},
			want: "Bearer token",
		},
		{
			name: "should return empty string",
			fields: fields{
				client:  clientWithCACert,
				options: options,
			},
			args: args{
				r: &http.Request{
					Header: http.Header{},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TokenVerifier{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			if got := tr.GetAuthenticationToken(tt.args.r); got != tt.want {
				t.Errorf("TokenVerifier.GetAuthenticationToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenVerifier_ValidateToken(t *testing.T) {
	type fields struct {
		client  http.Client
		options TokenVerifierOptions
	}
	type args struct {
		url       *url.URL
		clusterID string
		token     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return error",
			fields: fields{
				client:  clientWithCACert,
				options: options,
			},
			args: args{
				url: &url.URL{
					Scheme: "https",
					Host:   "localhost:8080",
				},
				clusterID: "clusterID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TokenVerifier{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			if err := tr.ValidateToken(tt.args.url, tt.args.clusterID, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("TokenVerifier.ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultTLSConfig(t *testing.T) {
	rootCAs, err := x509.SystemCertPool()
	var tlsConfig tls.Config
	if err != nil {
		log.Print("failed to load system cert pool")
	}
	type args struct {
		options TokenVerifierOptions
		rootCA  *x509.CertPool
	}
	tests := []struct {
		name    string
		args    args
		want    tls.Config
		wantErr bool
	}{
		{
			name: "should return tls config if CACertEnabled is false",
			args: args{
				options: optionsWithoutCACert,
				rootCA:  rootCAs,
			},
			want:    tlsConfig,
			wantErr: false,
		},
		{
			name: "should return error",
			args: args{
				options: options,
			},
			want:    tls.Config{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultTLSConfig(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultTLSConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
