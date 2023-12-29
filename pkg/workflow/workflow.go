package Workflow

import (
	"fmt"
	"time"

	"github.com/pothulapati/tostgres/pkg/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Tostgres struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func CreateTostgres(ctx workflow.Context, instance *Tostgres) error {
	activitiesContext := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			//MaximumAttempts: 3,
			BackoffCoefficient: 1,
		},
	})

	doActivities := activities.NewDoActivities()
	var dropletId int
	err := workflow.ExecuteActivity(activitiesContext, doActivities.SpinUpDroplet, instance.Name, instance.Region, "cncfhyderabad").Get(ctx, &dropletId)
	if err != nil {
		return fmt.Errorf("failed to spin up droplet: %w", err)
	}

	var dropletIp string
	err = workflow.ExecuteActivity(activitiesContext, doActivities.WaitForDroplet, dropletId).Get(ctx, &dropletIp)
	if err != nil {
		return fmt.Errorf("failed to wait for droplet: %w", err)
	}

	err = workflow.ExecuteActivity(activitiesContext, doActivities.UpdateDNS, "tostgres.cloud", instance.Name, dropletIp).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to update DNS: %w", err)
	}

	return nil
}
