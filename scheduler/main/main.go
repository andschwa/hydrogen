package main

import (
	"flag"
	"github.com/verizonlabs/sprint/scheduler"
	"os"
)

func main() {
	fs := flag.NewFlagSet("scheduler", flag.ExitOnError)

	config := new(scheduler.Configuration)
	config.Initialize(fs)

	fs.Parse(os.Args[1:])

	shutdown := make(chan struct{})

	sched := scheduler.NewScheduler(config, shutdown)
	controller := scheduler.NewController(sched, shutdown)
	handlers := scheduler.NewHandlers(sched)

	sched.Run(controller.GetSchedulerCtrl(), controller.BuildConfig(
		controller.BuildContext(),
		controller.BuildFrameworkInfo(config),
		sched.GetCaller(),
		shutdown,
		handlers,
	))
}
