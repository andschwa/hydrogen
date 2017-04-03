package api

import (
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task"
	"net/http"
	"testing"
)

// TODO think about how/where to keep these mocks as other tests start using them
// It sounds like a common pattern is to make a place for mocks and put them all in their own package(s).
// Since you cannot reference types in tests from tests in other packages people seem to be putting them in non-test files.

type mockScheduler struct{}

func (m *mockScheduler) Subscribe(chan *sched.Event) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Teardown() (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Accept(offerIds []*mesos_v1.OfferID, tasks []*mesos_v1.Offer_Operation, filters *mesos_v1.Filters) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Decline(offerIds []*mesos_v1.OfferID, filters *mesos_v1.Filters) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Revive() (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Kill(taskId *mesos_v1.TaskID, agentid *mesos_v1.AgentID) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Shutdown(execId *mesos_v1.ExecutorID, agentId *mesos_v1.AgentID) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Acknowledge(agentId *mesos_v1.AgentID, taskId *mesos_v1.TaskID, uuid []byte) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Reconcile(tasks []*mesos_v1.TaskInfo) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Message(agentId *mesos_v1.AgentID, executorId *mesos_v1.ExecutorID, data []byte) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) SchedRequest(resources []*mesos_v1.Request) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Suppress() (*http.Response, error) {
	return new(http.Response), nil
}

type mockTaskManager struct{}

func (m *mockTaskManager) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m *mockTaskManager) Delete(*mesos_v1.TaskInfo) {

}

func (m *mockTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m *mockTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m *mockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m *mockTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) {

}

func (m *mockTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{
		{},
	}, nil
}

func (m *mockTaskManager) TotalTasks() int {
	return 0
}

func (m *mockTaskManager) Tasks() *structures.ConcurrentMap {
	return structures.NewConcurrentMap()
}

type mockResourceManager struct{}

func (m *mockResourceManager) AddOffers(offers []*mesos_v1.Offer) {

}

func (m *mockResourceManager) HasResources() bool {
	return false
}

func (m *mockResourceManager) AddFilter(t *mesos_v1.TaskInfo, filters []task.Filter) error {
	return nil
}

func (m *mockResourceManager) Assign(task *mesos_v1.TaskInfo) (*mesos_v1.Offer, error) {
	return &mesos_v1.Offer{}, nil
}

func (m *mockResourceManager) Offers() []*mesos_v1.Offer {
	return []*mesos_v1.Offer{
		{},
	}
}

type mockLogger struct{}

func (m *mockLogger) Emit(severity uint8, template string, args ...interface{}) {

}

type mockServerConfiguration struct{}

func (m *mockServerConfiguration) Cert() string {
	return ""
}

func (m *mockServerConfiguration) Key() string {
	return ""
}

func (m *mockServerConfiguration) Port() int {
	return 0
}

func (m *mockServerConfiguration) Path() string {
	return ""
}

func (m *mockServerConfiguration) Protocol() string {
	return "http"
}

func (m *mockServerConfiguration) Server() *http.Server {
	return &http.Server{}
}

func (m *mockServerConfiguration) TLS() bool {
	return false
}

var c = new(mockServerConfiguration)
var s = new(mockScheduler)
var tm = new(mockTaskManager)
var r = new(mockResourceManager)
var h = http.NewServeMux()
var v = "test"
var l = new(mockLogger)

// Ensures all components are set correctly when creating the API server.
func TestNewApiServer(t *testing.T) {
	t.Parallel()

	srv := NewApiServer(c, s, tm, r, h, v, l)
	if srv.cfg != c || srv.sched != s || srv.taskMgr != tm || srv.resourceMgr != r ||
		srv.mux != h || srv.version != v || srv.logger != l {

		t.Fatal("API does not contain the correct components")
	}

}

// Checks if our internal handlers are attached correctly.
func TestApiServer_Handle(t *testing.T) {
	t.Parallel()

	srv := NewApiServer(c, s, tm, r, h, v, l)
	handles := map[string]http.HandlerFunc{
		"test1": func(w http.ResponseWriter, r *http.Request) {},
		"test2": func(w http.ResponseWriter, r *http.Request) {},
	}
	srv.setHandlers(handles)

	h := srv.Handle()
	if len(h) != len(handles) {
		t.Fatal("Not all handlers were applied correctly")
	}
}