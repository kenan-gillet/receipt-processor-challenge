#/bin/bash

# go run . &

for t in $(find examples -type f | sort)
    do
        id=$(curl -s -X POST -d @"$t" localhost:8080/receipts/process -H "Accept: application/json" | jq -r .id)
        echo  "for $t :: id=$id"
        curl localhost:8080/receipts/$id/points
    done

# pkill  receipt-processor
