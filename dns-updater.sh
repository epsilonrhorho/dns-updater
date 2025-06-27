#!/usr/bin/env bash
set -euo pipefail

CONFIG_FILE=${1:-/usr/local/etc/dns-updater.conf}
UPDATE_INTERVAL=${UPDATE_INTERVAL:-300}
STORAGE_PATH=${STORAGE_PATH:-/tmp/dns-updater}
mkdir -p "$STORAGE_PATH"

get_ip() {
  curl -sf https://api.ipify.org
}

update_cf() {
  local name=$1 zone=$2 ttl=$3 token=$4 ip=$5
  local api="https://api.cloudflare.com/client/v4"
  local record_id
  record_id=$(curl -sf -H "Authorization: Bearer $token" -H "Content-Type: application/json" \
    "$api/zones/$zone/dns_records?type=A&name=$name" | jq -r '.result[0].id')
  if [[ -z "$record_id" || "$record_id" == "null" ]]; then
    curl -sf -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" \
      --data "{\"type\":\"A\",\"name\":\"$name\",\"content\":\"$ip\",\"ttl\":$ttl}" \
      "$api/zones/$zone/dns_records" >/dev/null
  else
    curl -sf -X PUT -H "Authorization: Bearer $token" -H "Content-Type: application/json" \
      --data "{\"type\":\"A\",\"name\":\"$name\",\"content\":\"$ip\",\"ttl\":$ttl}" \
      "$api/zones/$zone/dns_records/$record_id" >/dev/null
  fi
}

update_r53() {
  local name=$1 zone=$2 ttl=$3 profile=$4 ip=$5
  local batch
  batch=$(cat <<JSON
{ "Changes": [ { "Action": "UPSERT", "ResourceRecordSet": { "Name": "$name", "Type": "A", "TTL": $ttl, "ResourceRecords": [ { "Value": "$ip" } ] } } ] }
JSON
)
  if [[ -n "$profile" && "$profile" != "-" ]]; then
    aws --profile "$profile" route53 change-resource-record-sets --hosted-zone-id "$zone" --change-batch "$batch" >/dev/null
  else
    aws route53 change-resource-record-sets --hosted-zone-id "$zone" --change-batch "$batch" >/dev/null
  fi
}

read_config() {
  mapfile -t CONFIG_LINES < <(grep -Ev '^\s*(#|$)' "$CONFIG_FILE")
}

main_loop() {
  while true; do
    ip=$(get_ip)
    for line in "${CONFIG_LINES[@]}"; do
      set -- $line
      name=$1 provider=$2 zone=$3 ttl=$4 cred=${5:-}
      file="$STORAGE_PATH/${name//./_}"
      last_ip=""
      [[ -f "$file" ]] && last_ip=$(cat "$file")
      if [[ "$ip" != "$last_ip" ]]; then
        if [[ "$provider" == "cloudflare" ]]; then
          update_cf "$name" "$zone" "$ttl" "$cred" "$ip"
        elif [[ "$provider" == "route53" ]]; then
          update_r53 "$name" "$zone" "$ttl" "$cred" "$ip"
        fi
        echo "$ip" > "$file"
        echo "$(date -Iseconds) Updated $name -> $ip"
      fi
    done
    sleep "$UPDATE_INTERVAL"
  done
}

read_config
main_loop
