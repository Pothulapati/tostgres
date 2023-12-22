package Workflow

import (
	"fmt"
	"time"

	"github.com/pothulapati/tostgres/pkg/activities"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, token string, dropletName string, region string, size string, image string, domainName string, recordName string) error {
	activitiesContext := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	var doActivities *activities.DoActivities

	var dropletIP string
	err := workflow.ExecuteActivity(activitiesContext, doActivities.SpinUpDroplet, token, dropletName, region, size, image).Get(ctx, &dropletIP)
	if err != nil {
		return fmt.Errorf("failed to spin up droplet: %w", err)
	}

	err = workflow.ExecuteActivity(activitiesContext, doActivities.UpdateDNS, token, domainName, recordName, dropletIP).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to update DNS: %w", err)
	}

	return nil
}
