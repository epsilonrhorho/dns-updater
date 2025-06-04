package main

import (
	"context"
	"log"
	"os"

	"github.com/epsilonrhorho/dns-updater/ipify"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func main() {
	client := ipify.NewClient(nil)
	ip, err := client.GetIP(context.Background())
	if err != nil {
		log.Fatalf("failed to get IP: %v", err)
	}
	log.Println("Public IP:", ip)

	zoneID := os.Getenv("HOSTED_ZONE_ID")
	if zoneID == "" {
		log.Fatal("HOSTED_ZONE_ID environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	r53 := route53.NewFromConfig(cfg)

	_, err = r53.ChangeResourceRecordSets(context.Background(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name:            aws.String("raspberrypi.epsilonrhorho.club"),
						Type:            types.RRTypeA,
						TTL:             aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{{Value: aws.String(ip)}},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to update Route53 record: %v", err)
	}

	log.Println("Route53 A record updated")
}
