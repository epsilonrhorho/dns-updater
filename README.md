# dns-updater

`dns-updater` retrieves the machine's public IP address and updates DNS A records. It supports multiple DNS providers including AWS Route53 and Cloudflare, and can manage multiple hostnames simultaneously.

## Usage

```bash
dns-updater [-c config_file]
```

**Options:**
- `-c config_file` - Path to configuration file (default: `/usr/local/etc/dns-updater.yaml`)

## Configuration

The application now uses YAML configuration instead of environment variables to support multiple DNS records.

### Configuration File

Create a configuration file (default: `/usr/local/etc/dns-updater.yaml`) or specify a custom path using the `-c` flag:

```yaml
update_interval: 2m
storage_path: /tmp/dns-updater

records:
  foo.example.com:
    provider: cloudflare
    ttl: 60s
    cf_api_token: your_cloudflare_api_token_here
    
  bar.example.com:
    provider: route53
    ttl: 300s
    aws_access_key_id: your_aws_access_key_id
    aws_secret_key: your_aws_secret_access_key
    aws_region: us-east-1
    
  baz.subdomain.example.com:
    provider: cloudflare
    ttl: 120s
    cf_email: your_email@example.com
    cf_api_key: your_cloudflare_global_api_key
```

### Configuration Options

**Global Settings:**
- `update_interval` – how often to check for IP changes (default: `2m`)
- `storage_path` – base directory to persist last seen IP addresses (default: `/tmp/dns-updater`)

**Per-Record Settings:**
- `provider` – DNS provider (`route53` or `cloudflare`)
- `ttl` – DNS record TTL (default: `60s`)

**Note:** The DNS zone is automatically extracted from the record name. For example, `foo.example.com` will use zone `example.com`. Record names must have at least 3 DNS labels (e.g., `host.domain.tld`).

**AWS Route53 Settings:**
- `aws_access_key_id` – AWS access key ID
- `aws_secret_key` – AWS secret access key  
- `aws_region` – AWS region

**Cloudflare Settings:**
- `cf_api_token` – Cloudflare API token (recommended)
- `cf_email` + `cf_api_key` – Legacy authentication method

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

The manifests in `k8s/` run `dns-updater` continuously. You'll need to create a ConfigMap with your YAML configuration.

### Create configuration:
```bash
kubectl create configmap dns-updater-config --from-file=config.yaml
```

### Create secrets for sensitive data:
Instead of putting credentials directly in the YAML file, you can use Kubernetes secrets and reference them in your configuration:

```bash
kubectl create secret generic dns-updater-secrets \
  --from-literal=cf-api-token=<your-cloudflare-token> \
  --from-literal=aws-access-key-id=<your-aws-key> \
  --from-literal=aws-secret-key=<your-aws-secret>
```

### Deploy:
```bash
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/deployment.yaml
```

Note: You'll need to update the Kubernetes manifests to mount the ConfigMap and use the new YAML configuration format.

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

3. **Create the configuration and storage directories:**
   ```bash
   sudo mkdir -p /etc/dns-updater /var/lib/dns-updater
   sudo chown dns-updater:dns-updater /var/lib/dns-updater
   ```

4. **Create your configuration file:**
   ```bash
   sudo cp config.yaml.example /etc/dns-updater/config.yaml
   sudo chown dns-updater:dns-updater /etc/dns-updater/config.yaml
   sudo chmod 600 /etc/dns-updater/config.yaml
   ```
   
   Or place it in the default location:
   ```bash
   sudo mkdir -p /usr/local/etc
   sudo cp config.yaml.example /usr/local/etc/dns-updater.yaml
   sudo chown dns-updater:dns-updater /usr/local/etc/dns-updater.yaml
   sudo chmod 600 /usr/local/etc/dns-updater.yaml
   ```
   
   Edit the configuration file with your DNS records and credentials.

5. **Install the service file:**
   ```bash
   sudo cp dns-updater.service /etc/systemd/system/
   ```

6. **Update the service file to use the configuration file:**
   ```bash
   sudo systemctl edit dns-updater.service
   ```
   
   Update the ExecStart line to specify the configuration file:
   ```ini
   [Service]
   ExecStart=/usr/local/bin/dns-updater -c /usr/local/etc/dns-updater.yaml
   ```

7. **Enable and start the service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable dns-updater.service
   sudo systemctl start dns-updater.service
   ```

8. **Check the service status:**
   ```bash
   sudo systemctl status dns-updater.service
   sudo journalctl -u dns-updater.service -f
   ```
