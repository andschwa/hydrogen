// TODO if these are to be unit tests then we should have it in the same package as api.go.
// Remove these comments if the plan is for these to be functional tests.
// Also remove the _test suffix from the filename so that this isn't picked up with Go's testing command.
package test

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	"mesos-framework-sdk/structures"
	"net"
	"net/http"
	"os"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/api/response"
	"sprint/scheduler/events"
	tmanager "sprint/task/manager"
	"strings"
	"testing"
)

// Just see if our API is listening on 8081
func TestApiListen(t *testing.T) {
	_, err := net.Dial("tcp", ":"+os.Getenv("apiport"))
	if err != nil {
		t.Logf("API is unreachable on %v", os.Getenv("apiport"))
		t.FailNow()
	} else {
		t.Log("Connected")
	}
}

func TestApiDeploy(t *testing.T) {
	// Create our default task here.
	j := []byte(`{
	"name": "test application",
	"resources": {"cpus": 0.1, "mem": 128.0},
	"command": {"cmd": "echo test"},
	"container": {"image": "alpine:latest"},
	"healthcheck": {"endpoint": "localhost:8080"},
	"labels": [{"purpose": "Testing"}]}`)

	req, err := http.NewRequest("POST", "http://localhost:"+os.Getenv("apiport")+"/v1/api/deploy", bytes.NewBuffer(j))
	if err != nil {
		t.Fail()
	}
	req.Header.Set("Content-Type", "application/json")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Log(err.Error())
	}
	if m.Status != response.ACCEPTED {
		t.Fail()
	}
}

func SetupEnv() {
	os.Setenv("endpoint", "http://localhost:5050/api/v1/scheduler")
	os.Setenv("apiport", "8080")
	cert := flag.String("server.cert", "", "TLS certificate")
	key := flag.String("server.key", "", "TLS key")
	path := flag.String("server.executor.path", "executor", "Path to the executor binary")
	port := flag.Int("server.executor.port", 8081, "Executor server listen port")

	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize()
	logger := logging.NewDefaultLogger()
	executorSrv := file.NewExecutorServer(srvConfig, logger)
	apiSrv := api.NewApiServer(srvConfig, http.NewServeMux(), port, "v1", logger)

	// Parse here to catch flags defined in structures above.
	flag.Parse()

	log.Println("Starting executor server...")
	go executorSrv.Serve()

	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            &schedulerConfig.User,
		Name:            &schedulerConfig.Name,
		FailoverTimeout: &schedulerConfig.Failover,
		Checkpoint:      &schedulerConfig.Checkpointing,
		Role:            &schedulerConfig.Role,
		Hostname:        &schedulerConfig.Hostname,
		Principal:       &schedulerConfig.Principal,
	}

	eventChan := make(chan *mesos_v1_scheduler.Event)

	// TODO need to mock the storage stuff, otherwise creating the client fails.
	kv := etcd.NewClient(strings.Split(schedulerConfig.StorageEndpoints, ","), schedulerConfig.StorageTimeout)
	man := tmanager.NewTaskManager(structures.NewConcurrentMap())
	c := client.NewClient(schedulerConfig.MesosEndpoint, logger)
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger)
	r := manager.NewDefaultResourceManager()
	e := eventcontroller.NewSprintEventController(s, man, r, eventChan, kv, logger)

	log.Println("Starting API server...")
	go apiSrv.RunAPI(e, nil) // TODO mock API for unit tests.

	go e.Run() // TODO mock what we can so that we don't really try and make a connection to Mesos.
}

func TestMain(m *testing.M) {
	SetupEnv()
	os.Exit(m.Run())
}
