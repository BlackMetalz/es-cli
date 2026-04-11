#!/bin/bash
# Seed demo data into Elasticsearch
# Creates demo-1, demo-2, demo-3 indices with ~50MB each

ES_URL="${ES_URL:-http://localhost:9200}"
ES_USER="${ES_USER:-elastic}"
ES_PASS="${ES_PASS:-elastic}"
CURL="curl -s -u ${ES_USER}:${ES_PASS}"

INDICES=("demo-1" "demo-2" "demo-3")
TARGET_SIZE_MB=50
# Each doc is ~1KB, so ~51200 docs per index for ~50MB
DOCS_PER_INDEX=51200
BATCH_SIZE=5000

echo "Seeding data to ${ES_URL}..."

for INDEX in "${INDICES[@]}"; do
    echo ""
    echo "=== Creating index: ${INDEX} ==="

    # Delete if exists
    $CURL -XDELETE "${ES_URL}/${INDEX}" 2>/dev/null
    echo ""

    # Create index
    $CURL -XPUT "${ES_URL}/${INDEX}" -H 'Content-Type: application/json' -d '{
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 0,
            "refresh_interval": "30s"
        }
    }'
    echo ""

    TOTAL=0
    while [ $TOTAL -lt $DOCS_PER_INDEX ]; do
        # Build bulk request
        BULK=""
        REMAINING=$((DOCS_PER_INDEX - TOTAL))
        CURRENT_BATCH=$BATCH_SIZE
        if [ $REMAINING -lt $BATCH_SIZE ]; then
            CURRENT_BATCH=$REMAINING
        fi

        for ((i=0; i<CURRENT_BATCH; i++)); do
            # Generate ~1KB of random data per document
            TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")
            RANDOM_NUM=$((RANDOM % 1000000))
            RANDOM_FLOAT=$(echo "scale=2; $RANDOM / 32767 * 100" | bc 2>/dev/null || echo "$((RANDOM % 10000)).$((RANDOM % 100))")
            # Pad with random text to reach ~1KB per doc
            PADDING=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9 ' | head -c 700)
            BULK+=$'{"index":{}}\n'
            BULK+="{\"timestamp\":\"${TIMESTAMP}\",\"value\":${RANDOM_NUM},\"metric\":${RANDOM_FLOAT},\"status\":\"active\",\"host\":\"server-$((RANDOM % 10))\",\"message\":\"${PADDING}\"}"$'\n'
        done

        RESPONSE=$($CURL -XPOST "${ES_URL}/${INDEX}/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary "$BULK" 2>&1)

        TOTAL=$((TOTAL + CURRENT_BATCH))
        PERCENT=$((TOTAL * 100 / DOCS_PER_INDEX))
        printf "\r  [${INDEX}] Progress: %d/%d docs (%d%%)" $TOTAL $DOCS_PER_INDEX $PERCENT
    done

    # Force refresh
    $CURL -XPOST "${ES_URL}/${INDEX}/_refresh" > /dev/null 2>&1

    # Show index size
    echo ""
    SIZE=$($CURL "${ES_URL}/_cat/indices/${INDEX}?h=pri.store.size" 2>/dev/null | tr -d '[:space:]')
    echo "  [${INDEX}] Done! Size: ${SIZE}"
done

echo ""
echo "=== Seed complete ==="
$CURL "${ES_URL}/_cat/indices/demo-*?v&h=health,status,index,pri,rep,docs.count,pri.store.size"
echo ""
