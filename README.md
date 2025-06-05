# dns-updater

`dns-updater` retrieves the machine's public IP address and updates a Route53 A record. The specific record to update is supplied via an environment variable so the program can be used for any host name.

## Required environment variables

- `HOSTED_ZONE_ID` – the ID of the Route53 hosted zone containing the record.
- `RECORD_NAME` – the fully qualified DNS record name to update.
- `AWS_ACCESS_KEY_ID` – your AWS access key ID used to authenticate with Route53.
- `AWS_SECRET_ACCESS_KEY` – your AWS secret access key.

## IAM policy requirements

The AWS credentials used must permit `ChangeResourceRecordSets` on the hosted zone. Below is an example policy you can attach to the user or role executing this program:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets",
        "route53:GetHostedZone"
      ],
      "Resource": "arn:aws:route53:::hostedzone/<HOSTED_ZONE_ID>"
    }
  ]
}
```

Replace `<HOSTED_ZONE_ID>` with your zone ID. The user or role must also have permission to retrieve changes if desired, e.g. `route53:GetChange`.
