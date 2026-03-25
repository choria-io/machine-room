#!/bin/bash

set -e

if [ -f /configuration/issuer/issuer.seed ]
then
  echo "Issuer exist, skipping"
  exit 0
fi

log() {
	echo
	echo ">>>>>>>"
	echo $1
	echo "<<<<<<<"
	echo
}

log "Creating Organization Issuer"
choria jwt keys /configuration/issuer/issuer.seed /configuration/issuer/issuer.public

log "Creating provisioner"
choria jwt keys /configuration/provisioner/signer.seed /configuration/provisioner/signer.public
choria jwt client /configuration/provisioner/signer.jwt provisioner_signer /configuration/issuer/issuer.seed --public-key "$(cat /configuration/provisioner/signer.public)" --server-provisioner --validity 365d --issuer
ls -l /setup/templates/provisioner/
cp -v /setup/templates/provisioner/choria.cfg /configuration/provisioner/
cp -v /setup/templates/provisioner/customers.json /configuration/provisioner/
cat /setup/templates/provisioner/provisioner.yaml|sed -e "s.ISSUER.$(cat /configuration/issuer/issuer.public)." > /configuration/provisioner/provisioner.yaml
cat /setup/templates/provisioner/helper.rb|sed -e "s.ISSUER.$(cat /configuration/issuer/issuer.public)." > /configuration/provisioner/helper.rb
chmod a+x /configuration/provisioner/helper.rb

log "Creating provisioner broker"
choria jwt keys /configuration/broker/broker.seed /configuration/broker/broker.public
choria jwt server /configuration/broker/broker.jwt broker.backend.saas.local "$(cat /configuration/broker/broker.public)" /configuration/issuer/issuer.seed --org choria --collectives choria --subjects 'choria.node_metadata.>'
openssl genrsa -out /configuration/broker/private.key 2048
openssl req -new -x509 -sha256 -key /configuration/broker/private.key -out /configuration/broker/public.crt -days 365 -subj "/O=Saas/CN=provision.backend.saas.local" -addext "subjectAltName = DNS:provision.backend.saas.local"
cat /setup/templates/broker/broker.conf|sed -e "s.ISSUER.$(cat /configuration/issuer/issuer.public)." > /configuration/broker/broker.conf

chown -R choria:choria /configuration

log "Creating customer provisioning jwt"
choria jwt prov /configuration/customer/tools/provisioning.jwt /configuration/issuer/issuer.seed --token s3cret --urls nats://provision-broker.backend.saas.local:4222 --default --protocol-v2 --insecure --update --validity 365d --extensions '{"customer":"one", "role":"tools"}'
choria jwt prov /configuration/customer/app1/provisioning.jwt /configuration/issuer/issuer.seed --token s3cret --urls nats://provision-broker.backend.saas.local:4222 --default --protocol-v2 --insecure --update --validity 365d --extensions '{"customer":"one", "role":"app1"}'
choria jwt prov /configuration/customer/app2/provisioning.jwt /configuration/issuer/issuer.seed --token s3cret --urls nats://provision-broker.backend.saas.local:4222 --default --protocol-v2 --insecure --update --validity 365d --extensions '{"customer":"one", "role":"app2"}'
choria jwt prov /configuration/customer/app3/provisioning.jwt /configuration/issuer/issuer.seed --token s3cret --urls nats://provision-broker.backend.saas.local:4222 --default --protocol-v2 --insecure --update --validity 365d --extensions '{"customer":"one", "role":"app3"}'

log "Setting up SaaS NATS"
cp /setup/templates/saas-nats/* /configuration/saas-nats/

log "Setting up RedPanda Connect"
cp /setup/templates/nodes_processor/nodes.yaml /configuration/redpanda/