# dns-updater

`dns-updater` is a small Bash script that periodically checks the machine's public IP address and updates DNS A records. It supports AWS Route53 and Cloudflare.

## Usage

```bash
./dns-updater.sh [config_file]
```

- `config_file` – path to the configuration file. Defaults to `/usr/local/etc/dns-updater.conf`.
- `UPDATE_INTERVAL` – interval between checks in seconds (default `300`).
- `STORAGE_PATH` – directory to store the last seen IP (default `/tmp/dns-updater`).

## Configuration

The configuration file is a simple whitespace separated format. Each non-comment line describes one DNS record:

```
<record_name> <provider> <zone_id> <ttl> [credential]
```

- `provider` – `cloudflare` or `route53`.
- `zone_id` – Cloudflare zone ID or Route53 hosted zone ID.
- `ttl` – TTL in seconds.
- `credential` – Cloudflare API token or AWS profile name (use `-` to use the default AWS configuration).

Example `config.conf`:

```
# Cloudflare record
foo.example.com cloudflare cf_zone_id_here 60 cf_api_token_here

# Route53 record
bar.example.com route53 r53_zone_id_here 300 default
```

The script uses the [ipify](https://www.ipify.org) service to discover the public IP.

## Docker

A simple Docker image is provided. Build and run:

```bash
docker build -t dns-updater .
docker run -v $(pwd)/config.conf:/etc/dns-updater/config dns-updater
```

## Kubernetes

The manifests in `k8s/` run the script continuously. Mount your configuration file using a ConfigMap and provide a writable volume for the storage path.

## Systemd service

The `dns-updater.service` file can be installed on Linux systems to run the script as a service. Place your configuration at `/etc/dns-updater/config` and enable the service:

```bash
sudo systemctl enable dns-updater.service
sudo systemctl start dns-updater.service
```
