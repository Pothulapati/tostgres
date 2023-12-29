package activities

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type DoActivities struct {
	token string
}

func NewDoActivities() *DoActivities {
	token := os.Getenv("DO_TOKEN")
	return &DoActivities{
		token: token,
	}
}

// SpinUpDroplet spins up a DigitalOcean droplet
func (d *DoActivities) SpinUpDroplet(dropletName string, region, password string) (int, error) {
	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: d.token},
	))

	client := godo.NewClient(oauthClient)

	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: region,
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-23-10-x64",
		},
		UserData: fmt.Sprintf(`#!/bin/bash

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Run Postgres as a Docker container
docker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=%s -e POSTGRES_DB=%s -e POSTGRES_PASSWORD=%s postgres`, dropletName, dropletName, password),
	}

	newDroplet, _, err := client.Droplets.Create(ctx, createRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to create droplet: %w", err)
	}

	return newDroplet.ID, nil
}

// UpdateDNS updates the DNS record to point to the given IP address
func (d *DoActivities) UpdateDNS(domainName string, recordName string, ipAddress string) error {
	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: d.token},
	))

	client := godo.NewClient(oauthClient)

	domain, _, err := client.Domains.Get(ctx, domainName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	createRequest := &godo.DomainRecordEditRequest{
		Type: "A",
		Name: recordName,
		Data: ipAddress,
		TTL:  35,
	}

	_, _, err = client.Domains.CreateRecord(ctx, domain.Name, createRequest)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// waitForDroplet waits until the droplet is ready
func (d *DoActivities) WaitForDroplet(ctx context.Context, dropletID int) (string, error) {
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: d.token},
	))

	client := godo.NewClient(oauthClient)

	var ipv4 string
	for {
		droplet, _, err := client.Droplets.Get(ctx, dropletID)
		if err != nil {
			return "", fmt.Errorf("failed to get droplet: %w", err)
		}

		ipv4, err = droplet.PublicIPv4()
		if err != nil {
			return "", err
		}

		if droplet.Status == "active" && err == nil {
			break
		}

		time.Sleep(time.Second * 10)
	}

	return ipv4, nil
}
