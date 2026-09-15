#!/bin/bash
set -euo pipefail

topics=(
  platform.smoke-test.v1
  media.uploaded.v1
  media.processed.v1
  media.processing-failed.v1
  identity.email-verification-requested.v1
  identity.password-recovery-requested.v1
  identity.security-audit.v1
)

for topic in "${topics[@]}"; do
  /opt/kafka/bin/kafka-topics.sh \
    --bootstrap-server kafka:29092 \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions 1 \
    --replication-factor 1
done
