FROM alpine:3.18
RUN apk add --no-cache bash curl jq aws-cli
COPY dns-updater.sh /usr/local/bin/dns-updater.sh
COPY config.conf.example /etc/dns-updater/config
ENTRYPOINT ["/usr/local/bin/dns-updater.sh", "/etc/dns-updater/config"]
