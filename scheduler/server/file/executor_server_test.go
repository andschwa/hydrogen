package file

import (
	"reflect"
	"sprint/scheduler/server"
	"testing"
)

// Mocked configuration
type mockConfiguration struct {
	cfg server.ServerConfiguration
}

func (m *mockConfiguration) Initialize() *server.ServerConfiguration {
	*m.cfg.Port() = 8081

	return &m.cfg
}

func (m *mockConfiguration) Cert() string {
	return m.cfg.Cert()
}

func (m *mockConfiguration) Key() string {
	return m.cfg.Key()
}

func (m *mockConfiguration) Port() *int {
	return m.cfg.Port()
}

var cfg server.Configuration = new(mockConfiguration).Initialize()

// Make sure we get the right type for our executor server.
func TestNewExecutorServer(t *testing.T) {
	t.Parallel()

	path := "executor"
	port := 8081
	cert := ""
	key := ""

	srv := NewExecutorServer(cfg)
	if reflect.TypeOf(srv) != reflect.TypeOf(new(executorServer)) {
		t.Fatal("Executor server is of the wrong type")
	}

	if srv.path != path {
		t.Fatal("Executor server path was not set correctly")
	}
	if srv.port != port {
		t.Fatal("Executor server port was not set correctly")
	}
	if srv.cert != cert {
		t.Fatal("Executor server certificate was not set correctly")
	}
	if srv.key != key {
		t.Fatal("Executor server key was not set correctly")
	}
}
