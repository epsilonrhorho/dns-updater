package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// Interface defines the behavior for updating DNS records.
type Interface interface {
	UpdateARecord(ctx context.Context, hostedZoneID, recordName, ip string, ttl int64) error
}

// Route53API defines the minimal interface needed for Route53 operations.
type Route53API interface {
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
}

// Client wraps the AWS Route53 client.
type Client struct {
	client Route53API
}

// NewClient creates a new Route53 client wrapper.
func NewClient(client Route53API) *Client {
	return &Client{client: client}
}

// UpdateARecord updates an A record in Route53.
func (c *Client) UpdateARecord(ctx context.Context, hostedZoneID, recordName, ip string, ttl int64) error {
	_, err := c.client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name:            aws.String(recordName),
						Type:            types.RRTypeA,
						TTL:             aws.Int64(ttl),
						ResourceRecords: []types.ResourceRecord{{Value: aws.String(ip)}},
					},
				},
			},
		},
	})
	return err
}