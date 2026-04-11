#!/bin/bash
# Generate N empty indices with prefix demo-index-num-X
# Usage: ./scripts/generate-indices.sh [NUM]

ES_URL="${ES_URL:-http://localhost:9200}"
ES_USER="${ES_USER:-elastic}"
ES_PASS="${ES_PASS:-elastic}"
NUM="${1:-100}"
CURL="curl -s -u ${ES_USER}:${ES_PASS}"

echo "Creating ${NUM} indices at ${ES_URL}..."

for ((i=1; i<=NUM; i++)); do
    INDEX="demo-index-num-${i}"
    $CURL -XPUT "${ES_URL}/${INDEX}" -H 'Content-Type: application/json' -d '{
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 0
        }
    }' > /dev/null 2>&1
    printf "\r  Created %d/%d" $i $NUM
done

echo ""
echo "=== Done! ==="
$CURL "${ES_URL}/_cat/indices/demo-index-num-*?v&h=health,status,index,pri,rep,docs.count,pri.store.size&s=index"
echo ""
