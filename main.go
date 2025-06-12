package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/epsilonrhorho/dns-updater/ipify"
	"github.com/epsilonrhorho/dns-updater/route53"
	"github.com/epsilonrhorho/dns-updater/service"
	"github.com/epsilonrhorho/dns-updater/storage"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
)

type Config struct {
	HostedZoneID string `envconfig:"HOSTED_ZONE_ID" required:"true"`
	RecordName   string `envconfig:"RECORD_NAME" required:"true"`
	AccessKeyID  string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretKey    string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	Region       string `envconfig:"AWS_REGION" required:"true"`
	StoragePath  string `envconfig:"STORAGE_PATH" required:"true"`
	UpdateInterval time.Duration `envconfig:"UPDATE_INTERVAL" default:"2m"`
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

	awsRoute53Client := awsroute53.NewFromConfig(cfg)
	route53Client := route53.NewClient(awsRoute53Client)
	ipClient := ipify.NewClient(nil)
	storageClient := storage.NewFileStorage(c.StoragePath)

	serviceConfig := service.Config{
		HostedZoneID: c.HostedZoneID,
		RecordName:   c.RecordName,
		TTL:          60,
	}

	dnsService := service.New(route53Client, ipClient, storageClient, serviceConfig, c.UpdateInterval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dnsService.Run(ctx)
}
