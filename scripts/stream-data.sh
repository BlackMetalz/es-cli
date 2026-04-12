#!/bin/bash
# Continuously insert log-like data into a specified index
# Usage: ./scripts/stream-data.sh [INDEX] [INTERVAL_MS]
# Example: make stream-data INDEX=app-logs INTERVAL=500

ES_URL="${ES_URL:-http://localhost:9200}"
ES_USER="${ES_USER:-elastic}"
ES_PASS="${ES_PASS:-elastic}"
INDEX="${1:-app-logs}"
INTERVAL_MS="${2:-1000}"
CURL="curl -s -u ${ES_USER}:${ES_PASS}"
BATCH_SIZE=10

LEVELS=("INFO" "WARN" "ERROR" "DEBUG")
SERVICES=("api" "web" "worker" "scheduler" "auth")
MESSAGES_INFO=("Request completed" "Cache hit" "User logged in" "Health check passed" "Task completed" "Connection established")
MESSAGES_WARN=("High memory usage" "Slow query detected" "Retry attempt" "Rate limit approaching" "Disk usage above 80%")
MESSAGES_ERROR=("Connection refused" "Timeout after 30s" "Out of memory" "Permission denied" "Service unavailable" "Null pointer exception")
MESSAGES_DEBUG=("Processing request" "Checking cache" "Loading config" "Parsing payload" "Validating token")

# Create index if not exists
$CURL -XPUT "${ES_URL}/${INDEX}" -H 'Content-Type: application/json' -d '{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "properties": {
            "@timestamp": {"type": "date"},
            "level": {"type": "keyword"},
            "service": {"type": "keyword"},
            "message": {"type": "text"},
            "host": {"type": "keyword"},
            "status_code": {"type": "integer"},
            "duration_ms": {"type": "integer"}
        }
    }
}' 2>/dev/null

echo "Streaming data to ${ES_URL}/${INDEX} every ${INTERVAL_MS}ms (Ctrl+C to stop)..."
echo ""

COUNT=0
while true; do
    BULK=""
    for ((i=0; i<BATCH_SIZE; i++)); do
        # Random level (weighted: 60% INFO, 15% WARN, 15% ERROR, 10% DEBUG)
        RAND=$((RANDOM % 100))
        if [ $RAND -lt 60 ]; then
            LEVEL="INFO"
            MSG="${MESSAGES_INFO[$((RANDOM % ${#MESSAGES_INFO[@]}))]}"
            STATUS=$((200 + (RANDOM % 2) * 4))
        elif [ $RAND -lt 75 ]; then
            LEVEL="WARN"
            MSG="${MESSAGES_WARN[$((RANDOM % ${#MESSAGES_WARN[@]}))]}"
            STATUS=$((400 + RANDOM % 10))
        elif [ $RAND -lt 90 ]; then
            LEVEL="ERROR"
            MSG="${MESSAGES_ERROR[$((RANDOM % ${#MESSAGES_ERROR[@]}))]}"
            STATUS=$((500 + RANDOM % 4))
        else
            LEVEL="DEBUG"
            MSG="${MESSAGES_DEBUG[$((RANDOM % ${#MESSAGES_DEBUG[@]}))]}"
            STATUS=200
        fi

        SERVICE="${SERVICES[$((RANDOM % ${#SERVICES[@]}))]}"
        HOST="host-$((RANDOM % 5 + 1))"
        DURATION=$((RANDOM % 2000))
        TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")

        BULK+=$'{"index":{}}\n'
        BULK+="{\"@timestamp\":\"${TIMESTAMP}\",\"level\":\"${LEVEL}\",\"service\":\"${SERVICE}\",\"message\":\"${MSG}\",\"host\":\"${HOST}\",\"status_code\":${STATUS},\"duration_ms\":${DURATION}}"$'\n'
    done

    $CURL -XPOST "${ES_URL}/${INDEX}/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary "$BULK" > /dev/null 2>&1

    COUNT=$((COUNT + BATCH_SIZE))
    printf "\r  [%s] %d docs inserted" "$INDEX" "$COUNT"

    # Sleep for interval
    sleep $(echo "scale=3; ${INTERVAL_MS}/1000" | bc 2>/dev/null || echo "1")
done
