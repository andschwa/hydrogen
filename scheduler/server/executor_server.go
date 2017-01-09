package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
)

type executorServer struct {
	mux  *http.ServeMux
	path string
	port int
	tls  bool
	cert string
	key  string
}

// Returns a new instance of our server.
func NewExecutorServer(path string, port int, cert, key string) *executorServer {
	enableTls := cert != "" && key != ""

	return &executorServer{
		mux:  http.NewServeMux(),
		path: path,
		port: port,
		tls:  enableTls,
		cert: cert,
		key:  key,
	}
}

// Handler to serve the executor binary.
func (s *executorServer) executorHandlers(path string, tls bool) {
	s.mux.HandleFunc("/executor", s.executorBinary)
}

func (s *executorServer) executorBinary(w http.ResponseWriter, r *http.Request) {
	if s.tls {
		// Don't allow fallbacks to HTTP.
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
	http.ServeFile(w, r, s.path)
}

// Serve the executor over plain HTTP.
func (s *executorServer) Serve() {
	s.executorHandlers(s.path, s.tls)

	if s.tls {
		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(s.port),
			Handler: s.mux,
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

		log.Fatal(srv.ListenAndServeTLS(s.cert, s.key))
	} else {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.port), s.mux))
	}
}
