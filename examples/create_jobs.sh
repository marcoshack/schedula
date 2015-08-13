#!/bin/bash

SCHED_URI=${SCHED_URI:-http://localhost:8080/jobs/}
SCHED_CALLBACK=${SCHED_CALLBACK:-http://localhost:9292/sniffer}
SCHED_TOTAL=${SCHED_TOTAL:-1}
SCHED_DELTA=${SCHED_DELTA:-2}
SCHED_TIME=${SCHED_TIME:-$(expr $(date +%s) + $SCHED_DELTA)}
I_SCHED_START=$SECONDS
(set -o posix; set) |grep ^SCHED_.*

for i in `seq 1 $SCHED_TOTAL`; do
  curl $SCHED_URI -XPOST -H'Content-Type: application/json' -d  "{\"callbackURL\":\"$SCHED_CALLBACK\",\"schedule\":{\"format\":\"timestamp\",\"value\":\"$SCHED_TIME\"}}"

  if [[ $(expr $i % 100 ) == "0" ]]; then
    echo "$(date): $i/$SCHED_TOTAL (~$(expr $i / $(expr $SECONDS - $I_SCHED_START)) RPS)"
  fi
done
echo "done!"
