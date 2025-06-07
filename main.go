package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	Region       string `envconfig:"AWS_REGION" required:"true"`
	StoragePath  string `envconfig:"STORAGE_PATH" required:"true"`
}

func update(ctx context.Context, r53 *route53.Client, ipClient ipify.ClientInterface, c Config) error {
	ip, err := ipClient.GetIP(ctx)
	if err != nil {
		return err
	}

	prevData, err := os.ReadFile(c.StoragePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lastIP := strings.TrimSpace(string(prevData))

	if lastIP == ip {
		return nil
	}

	_, err = r53.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(c.HostedZoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name:            aws.String(c.RecordName),
						Type:            types.RRTypeA,
						TTL:             aws.Int64(60),
						ResourceRecords: []types.ResourceRecord{{Value: aws.String(ip)}},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.StoragePath, []byte(ip), 0600); err != nil {
		return err
	}

	log.Printf("Address of %s was updated to %s", c.RecordName, ip)
	return nil
}

func main() {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretKey, ""),
		),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	r53 := route53.NewFromConfig(cfg)

	ipClient := ipify.NewClient(nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		if err := update(ctx, r53, ipClient, c); err != nil {
			log.Printf("update failed: %v", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			log.Println("shutting down")
			return
		}
	}
}
