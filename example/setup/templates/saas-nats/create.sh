#!/bin/sh

export NATS_URL="nats://saas-nats.backend.saas.local:4222"
export NATS_USER="backend"
export NATS_PASSWORD="s3cret"

ready=""

while [ -z "${ready}" ]
do
  if nats account info
  then
    ready="yes"
  else
    echo "Waiting for server to be ready"
    sleep 1
  fi
done

nats stream add --config /machine-room/events.json
nats stream add --config /machine-room/nodes.json
nats stream add --config /machine-room/submit.json

NATS_USER="cust_one_admin"
NATS_PASSWORD="s3cret"

nats kv add CONFIG
nats kv put CONFIG machines "$(cat /machine-room/plugins.json)"