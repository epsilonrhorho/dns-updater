# dns-updater

`dns-updater` retrieves the machine's public IP address and updates the Route53 A record for `raspberrypi.epsilonrhorho.club`.

## Required environment variables

- `HOSTED_ZONE_ID` – the ID of the Route53 hosted zone containing the record.
- AWS credentials and region variables recognised by the AWS SDK such as `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_REGION` must also be configured so the program can authenticate with Route53.

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
