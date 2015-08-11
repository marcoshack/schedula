#!/bin/bash

for i in `seq 1 1`; do
  curl -XPOST -d "{\"callbackURL\":\"http://localhost:9999/callback\",\"schedule\":{\"format\":\"timestamp\",\"value\":\"$(expr $(date +%s) + 10)\"}}" -H'Content-Type: application/json' http://localhost:8080/jobs/
done
echo "done!"
