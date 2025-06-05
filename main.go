package main

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"

	"github.com/epsilonrhorho/dns-updater/ipify"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type Config struct {
	HostedZoneID string `envconfig:"HOSTED_ZONE_ID" required:"true"`
	RecordName   string `envconfig:"RECORD_NAME" required:"true"`
	AccessKeyID  string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretKey    string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

func main() {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}

	client := ipify.NewClient(nil)
	ip, err := client.GetIP(context.Background())
	if err != nil {
		log.Fatalf("failed to get IP: %v", err)
	}
	log.Println("Public IP:", ip)

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretKey, ""),
		),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	r53 := route53.NewFromConfig(cfg)

	_, err = r53.ChangeResourceRecordSets(context.Background(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(c.HostedZoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name:            aws.String(c.RecordName),
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
