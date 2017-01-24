package server

import (
	"crypto/tls"
	"flag"
	"net/http"
)

type Configuration interface {
	Initialize() *ServerConfiguration
	Cert() string
	Key() string
	Protocol() string
	Server() *http.Server
	TLS() bool
}

// Configuration for the executor server.
type ServerConfiguration struct {
	cert   string
	key    string
	path   string
	server *http.Server
	tls    bool
}

// Applies values to the various configurations from user-supplied flags.
func (c *ServerConfiguration) Initialize() *ServerConfiguration {
	flag.StringVar(&c.cert, "server.cert", "", "TLS certificate")
	flag.StringVar(&c.key, "server.key", "", "TLS key")
	c.server = &http.Server{
		TLSConfig: &tls.Config{
			// Use only the most secure protocol version.
			MinVersion: tls.VersionTLS12,
			// Use very strong crypto curves.
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			// Use very strong cipher suites (order is important here!)
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required for HTTP/2 support.
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	}
	c.tls = c.cert != "" && c.key != ""

	return c
}

// Gets the path to the TLS certificate.
func (c *ServerConfiguration) Cert() string {
	return c.cert
}

// Gets the path to the TLS key.
func (c *ServerConfiguration) Key() string {
	return c.key
}

// Determines the protocol to be used.
func (c *ServerConfiguration) Protocol() string {
	if c.cert != "" && c.key != "" {
		return "https"
	} else {
		return "http"
	}
}

// Returns the custom HTTP server with TLS configuration.
func (c *ServerConfiguration) Server() *http.Server {
	return c.server
}

// Returns true if TLS is enabled and false if TLS is disabled.
func (c *ServerConfiguration) TLS() bool {
	return c.tls
}
