package main

import (
	"log"

	"github.com/pothulapati/tostgres/pkg/activities"
	tcWorkflow "github.com/pothulapati/tostgres/pkg/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "default", worker.Options{})

	w.RegisterWorkflow(tcWorkflow.CreateTostgres)
	doActivities := activities.NewDoActivities()
	w.RegisterActivity(doActivities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
