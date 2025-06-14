# dns-updater

`dns-updater` retrieves the machine's public IP address and updates a DNS A record. It supports multiple DNS providers including AWS Route53 and Cloudflare.

## Required environment variables

### General Configuration
- `DNS_PROVIDER` – the DNS provider to use (`route53` or `cloudflare`)
- `ZONE` – the DNS zone name (e.g., `example.com`)
- `RECORD_NAME` – the DNS record name to update (e.g., `home` for `home.example.com`)
- `STORAGE_PATH` – path to persist the last seen IP address between runs

### Optional Configuration
- `UPDATE_INTERVAL` – how often to check for IP changes (default: `2m`)
- `TTL` – DNS record TTL (default: `60s`)

### AWS Route53 Configuration (when `DNS_PROVIDER=route53`)
- `AWS_ACCESS_KEY_ID` – your AWS access key ID
- `AWS_SECRET_ACCESS_KEY` – your AWS secret access key
- `AWS_REGION` – the AWS region to use

### Cloudflare Configuration (when `DNS_PROVIDER=cloudflare`)
- `CF_API_TOKEN` – Cloudflare API token with Zone:Edit permissions

## Provider-specific setup

### AWS Route53 IAM policy requirements

The AWS credentials used must permit `ChangeResourceRecordSets` on the hosted zone. Below is an example policy you can attach to the user or role executing this program:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets",
        "route53:GetHostedZone",
        "route53:ListResourceRecordSets"
      ],
      "Resource": "arn:aws:route53:::hostedzone/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones"
      ],
      "Resource": "*"
    }
  ]
}
```

### Cloudflare API token setup

For Cloudflare, create an API token with the following permissions:
- Zone:Edit permissions for the specific zone
- Zone:Read permissions for the specific zone

The token should be scoped to only the zone(s) you want to update.

## Kubernetes deployment

The manifests in `k8s/` run `dns-updater` continuously. The application now
handles its own scheduling and updates the DNS record every two minutes.

### For AWS Route53:
Create a secret containing your AWS credentials:

```bash
kubectl create secret generic dns-updater-secret \
  --from-literal=AWS_ACCESS_KEY_ID=<your-key-id> \
  --from-literal=AWS_SECRET_ACCESS_KEY=<your-secret-key>
```

### For Cloudflare:
Create a secret containing your Cloudflare credentials:

```bash
kubectl create secret generic dns-updater-secret \
  --from-literal=CF_API_TOKEN=<your-api-token>
```

Then create the PersistentVolumeClaim and Deployment:

```bash
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/deployment.yaml
```

Edit `k8s/deployment.yaml` to fill in your `DNS_PROVIDER`, `ZONE`, `RECORD_NAME` and other required environment variables for your chosen provider.

## Systemd service deployment

For running `dns-updater` as a systemd service on Linux:

1. **Build and install the binary:**
   ```bash
   go build -o dns-updater
   sudo cp dns-updater /usr/local/bin/
   sudo chmod +x /usr/local/bin/dns-updater
   ```

2. **Create a dedicated user:**
   ```bash
   sudo useradd --system --no-create-home --shell /bin/false dns-updater
   ```

3. **Create the storage directory:**
   ```bash
   sudo mkdir -p /var/lib/dns-updater
   sudo chown dns-updater:dns-updater /var/lib/dns-updater
   ```

4. **Install the service file:**
   ```bash
   sudo cp dns-updater.service /etc/systemd/system/
   ```

5. **Edit the service file with your configuration:**
   ```bash
   sudo systemctl edit dns-updater.service
   ```
   
   Add your configuration in the override file:
   
   **DNS Settings (required for all providers):**
   ```ini
   [Service]
   Environment=ZONE=yourdomain.com
   Environment=RECORD_NAME=home
   ```
   
   **AWS Route53 Settings:**
   ```ini
   Environment=DNS_PROVIDER=route53
   Environment=AWS_ACCESS_KEY_ID=your-access-key-id
   Environment=AWS_SECRET_ACCESS_KEY=your-secret-access-key
   Environment=AWS_REGION=us-east-1
   ```
   
   **Cloudflare Settings:**
   ```ini
   Environment=DNS_PROVIDER=cloudflare
   Environment=CF_API_TOKEN=your-api-token
   ```

6. **Enable and start the service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable dns-updater.service
   sudo systemctl start dns-updater.service
   ```

7. **Check the service status:**
   ```bash
   sudo systemctl status dns-updater.service
   sudo journalctl -u dns-updater.service -f
   ```
