package activities

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type DoActivities struct {
}

// SpinUpDroplet spins up a DigitalOcean droplet
func (d *DoActivities) SpinUpDroplet(token string, dropletName string, region string, size string, image string) (string, error) {
	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	client := godo.NewClient(oauthClient)

	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: image,
		},
		UserData: `#cloud-config
		package_upgrade: true
		packages:
			- docker.io
		write_files:
			- path: /etc/systemd/system/postgres.service
				content: |
					[Unit]
					Description=PostgreSQL Container
					After=docker.service
					Requires=docker.service
					
					[Service]
					Restart=always
					ExecStartPre=-/usr/bin/docker stop postgres
					ExecStartPre=-/usr/bin/docker rm postgres
					ExecStart=/usr/bin/docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres
					ExecStop=/usr/bin/docker stop postgres
					
					[Install]
					WantedBy=multi-user.target
		runcmd:
			- systemctl daemon-reload
			- systemctl enable --now postgres`,
	}

	newDroplet, _, err := client.Droplets.Create(ctx, createRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create droplet: %w", err)
	}

	ipv4, err := newDroplet.PublicIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get droplet IP address: %w", err)
	}

	return ipv4, nil
}

// UpdateDNS updates the DNS record to point to the given IP address
func (d *DoActivities) UpdateDNS(token string, domainName string, recordName string, ipAddress string) error {
	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	client := godo.NewClient(oauthClient)

	domain, _, err := client.Domains.Get(ctx, domainName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	records, _, err := client.Domains.Records(ctx, domain.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to get domain records: %w", err)
	}

	var recordID int
	for _, record := range records {
		if record.Name == recordName {
			recordID = record.ID
			break
		}
	}

	if recordID == 0 {
		return fmt.Errorf("record not found")
	}

	updateRequest := &godo.DomainRecordEditRequest{
		Data: ipAddress,
	}

	_, _, err = client.Domains.EditRecord(ctx, domain.Name, recordID, updateRequest)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}
