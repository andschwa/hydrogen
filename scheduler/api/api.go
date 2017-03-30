package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/server"
	taskbuilder "mesos-framework-sdk/task"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"net/http"
	"os"
	"sprint/scheduler/api/response"
	"sprint/scheduler/events"
	"sprint/task/builder"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
	killEndpoint   = "/kill"
	updateEndpoint = "/update"
	statsEndpoint  = "/stats"
	retries        = 20
)

type ApiServer struct {
	cfg       server.Configuration
	port      *int
	mux       *http.ServeMux
	handle    map[string]http.HandlerFunc
	eventCtrl *events.SprintEventController
	version   string
	logger    logging.Logger
}

func NewApiServer(cfg server.Configuration, mux *http.ServeMux, port *int, version string, lgr *logging.DefaultLogger) *ApiServer {
	return &ApiServer{
		cfg:     cfg,
		port:    port,
		mux:     mux,
		version: version,
		logger:  lgr,
	}
}

//Getter to return our map of handles
func (a *ApiServer) Handle() map[string]http.HandlerFunc {
	return a.handle
}

//Set our default API handler routes here.
func (a *ApiServer) setDefaultHandlers() {
	a.handle = make(map[string]http.HandlerFunc, 5)
	a.handle[baseUrl+deployEndpoint] = a.deploy
	a.handle[baseUrl+statusEndpoint] = a.state
	a.handle[baseUrl+killEndpoint] = a.kill
	a.handle[baseUrl+updateEndpoint] = a.update
	a.handle[baseUrl+statsEndpoint] = a.stats
}

func (a *ApiServer) setHandlers(handles map[string]http.HandlerFunc) {
	for route, handle := range handles {
		a.handle[route] = handle
	}
}

func (a *ApiServer) setEventController(e *events.SprintEventController) {
	a.eventCtrl = e
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI(e *events.SprintEventController, handlers map[string]http.HandlerFunc) {
	if handlers != nil || len(handlers) != 0 {
		a.logger.Emit(logging.INFO, "Setting custom handlers.")
		a.setHandlers(handlers)
	} else {
		a.logger.Emit(logging.INFO, "Setting default handlers.")
		a.setDefaultHandlers()
	}

	a.setEventController(e)

	// Iterate through all methods and setup endpoints.
	for route, handle := range a.handle {
		a.mux.HandleFunc(route, handle)
	}

	if a.cfg.TLS() {
		a.cfg.Server().Handler = a.mux
		a.cfg.Server().Addr = ":" + strconv.Itoa(*a.port)
		if err := a.cfg.Server().ListenAndServeTLS(a.cfg.Cert(), a.cfg.Key()); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(":"+strconv.Itoa(*a.port), a.mux); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	}
}

// Deploys a given application from parsed JSON
func (a *ApiServer) deploy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}

			var m taskbuilder.ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}

			task, err := builder.Application(&m, a.logger)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:   response.FAILED,
					TaskName: task.GetName(),
					Message:  err.Error(),
				})
				return
			}

			// If we have any filters, let the resource manager know.
			if len(m.Filters) > 0 {
				if err := a.eventCtrl.ResourceManager().AddFilter(task, m.Filters); err != nil {
					json.NewEncoder(w).Encode(response.Deploy{
						Status:   response.FAILED,
						TaskName: task.GetName(),
						Message:  err.Error(),
					})
					return
				}

			}

			if err := a.eventCtrl.TaskManager().Add(task); err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:   response.FAILED,
					TaskName: task.GetName(),
					Message:  err.Error(),
				})
				return
			}
			a.eventCtrl.Scheduler().Revive()

			json.NewEncoder(w).Encode(response.Deploy{
				Status:   response.QUEUED,
				TaskName: task.GetName(),
			})
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}

}

func (a *ApiServer) update(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}

			var m taskbuilder.ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}

			// Check if this task already exists
			taskToKill, err := a.eventCtrl.TaskManager().Get(&m.Name)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:   response.FAILED,
					TaskName: taskToKill.GetName(),
					Message:  err.Error(),
				})
				return
			}

			task, err := builder.Application(&m, a.logger)

			a.eventCtrl.TaskManager().Set(sdkTaskManager.UNKNOWN, task)
			a.eventCtrl.Scheduler().Kill(taskToKill.GetTaskId(), taskToKill.GetAgentId())
			a.eventCtrl.Scheduler().Revive()

			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.UPDATE,
				Message: fmt.Sprintf("Updating %v", task.GetName()),
			})
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}
}

func (a *ApiServer) kill(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return
			}

			var m taskbuilder.KillJson
			err = json.Unmarshal(dec, &m)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}

			// Make sure we have a name to look up
			if m.Name != nil {
				// Look up task in task manager
				t, err := a.eventCtrl.TaskManager().Get(m.Name)
				if err != nil {
					json.NewEncoder(w).Encode(response.Kill{Status: response.NOTFOUND, TaskName: *m.Name})
					return
				}
				// Get all tasks in RUNNING state.
				running, _ := a.eventCtrl.TaskManager().GetState(mesos_v1.TaskState_TASK_RUNNING)
				// If we get an error, it means no tasks are currently in the running state.
				// We safely ignore this- the range over the empty list will be skipped regardless.

				// Check if our task is in the list of RUNNING tasks.
				for _, task := range running {
					// If it is, then send the kill signal.
					if task.GetName() == t.GetName() {
						// First Kill call to the mesos-master.
						_, err := a.eventCtrl.Scheduler().Kill(t.GetTaskId(), t.GetAgentId())
						if err != nil {
							// If it fails, try to kill it again.
							resp, err := a.eventCtrl.Scheduler().Kill(t.GetTaskId(), t.GetAgentId())
							if err != nil {
								// We've tried twice and still failed.
								// Send back an error message.
								json.NewEncoder(w).Encode(
									response.Kill{
										Status:   response.FAILED,
										TaskName: *m.Name,
										Message:  "Response Status to Kill: " + resp.Status,
									},
								)
								return
							}
						}
						// Our kill call has worked, delete it from the task queue.
						a.eventCtrl.TaskManager().Delete(t)
						// Response appropriately.
						json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
						return
					}
				}
				// If we get here, our task isn't in the list of RUNNING tasks.
				// Delete it from the queue regardless.
				// We run into this case if a task is flapping or unable to launch
				// or get an appropriate offer.
				a.eventCtrl.TaskManager().Delete(t)
				json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
				return
			}
			// If we get here, there was no name passed in and the kill function failed.
			json.NewEncoder(w).Encode(response.Kill{Status: response.FAILED, TaskName: *m.Name})
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}

}

// TODO (tim): Get state of mesos task and return it.
func (a *ApiServer) stats(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			_, err := a.eventCtrl.TaskManager().Get(&name)
			if err != nil {
				fmt.Fprintf(w, "Task not found, error %v", err.Error())
				return
			}
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}
}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			_, err := a.eventCtrl.TaskManager().Get(&name)
			if err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:  response.FAILED,
					Message: err.Error(),
				})
				return
			}
			queued, err := a.eventCtrl.TaskManager().GetState(sdkTaskManager.STAGING)
			if err != nil {
				a.logger.Emit(logging.INFO, err.Error())
			}

			for _, task := range queued {
				if task.GetName() == name {
					json.NewEncoder(w).Encode(response.Kill{Status: response.QUEUED, TaskName: name})
				}
			}

			json.NewEncoder(w).Encode(response.Kill{Status: response.LAUNCHED, TaskName: name})

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
