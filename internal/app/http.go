package app

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"os"
	"time"
)

type HTTPOptions struct {
	Timeout  time.Duration
	Insecure bool
	Debug    bool
	DebugOut io.Writer
}

func NewHTTPClient(opts HTTPOptions) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.Insecure || os.Getenv("SSL_CERT_FILE") != "" {
		cfg := &tls.Config{InsecureSkipVerify: opts.Insecure} //nolint:gosec // explicit CLI opt-out.
		if path := os.Getenv("SSL_CERT_FILE"); path != "" {
			pem, err := os.ReadFile(path)
			if err != nil {
				return nil, WrapError("Could not read SSL_CERT_FILE", "Check the PEM bundle path and try again.", err)
			}
			pool, err := x509.SystemCertPool()
			if err != nil || pool == nil {
				pool = x509.NewCertPool()
			}
			if !pool.AppendCertsFromPEM(pem) {
				return nil, NewError("Could not parse SSL_CERT_FILE", "Point SSL_CERT_FILE at a PEM CA bundle.")
			}
			cfg.RootCAs = pool
		}
		transport.TLSClientConfig = cfg
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: opts.Timeout, Transport: transport}
	if opts.Debug && opts.DebugOut != nil {
		client.Transport = debugTransport{next: transport, out: opts.DebugOut}
	}
	return client, nil
}

type debugTransport struct {
	next http.RoundTripper
	out  io.Writer
}

func (d debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_, _ = io.WriteString(d.out, req.Method+" "+req.URL.String()+"\n")
	for k, values := range req.Header {
		for _, value := range values {
			if k == "Authorization" {
				value = RedactHeaderValue(value)
			}
			_, _ = io.WriteString(d.out, k+": "+value+"\n")
		}
	}
	return d.next.RoundTrip(req)
}
