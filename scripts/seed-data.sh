#!/bin/bash
# Seed demo data into Elasticsearch
# Creates demo-1, demo-2, demo-3 indices with ~50MB each
#
# Usage:
#   ./seed-data.sh
#   ES_URL=http://remote:9200 PARALLEL=8 BATCH_SIZE=20000 ./seed-data.sh

ES_URL="${ES_URL:-http://localhost:9200}"
ES_USER="${ES_USER:-elastic}"
ES_PASS="${ES_PASS:-elastic}"
CURL="curl -s -u ${ES_USER}:${ES_PASS}"

BATCH_SIZE="${BATCH_SIZE:-10000}"
PARALLEL="${PARALLEL:-4}"
DOCS_PER_INDEX=51200

INDICES=("demo-1" "demo-2" "demo-3")

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

# ── Python generator ──────────────────────────────────────────────────────────

gen_docs() {
    python3 - "$1" <<'PY'
import sys, random, datetime, json

count   = int(sys.argv[1])
HOSTS   = [f"server-{i}" for i in range(10)]
STATUSES= ['active', 'idle', 'degraded', 'maintenance']
METRICS = ['cpu_usage', 'mem_usage', 'disk_io', 'net_throughput', 'request_rate']

# Reusable padding pool — avoid per-doc random generation
ALPHA   = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 '
POOL    = ''.join(random.choices(ALPHA, k=16384))

now     = int(datetime.datetime.utcnow().timestamp())
RANGE_S = 30 * 86400

out = []
for i in range(count):
    ts  = datetime.datetime.utcfromtimestamp(now - random.randint(0, RANGE_S))
    off = (i * 700) % (len(POOL) - 720)
    pad = POOL[off: off + 700 + (i % 20)]
    doc = {
        "timestamp": ts.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
        "value":     random.randint(0, 1000000),
        "metric":    round(random.uniform(0.0, 100.0), 2),
        "metric_name": random.choice(METRICS),
        "status":    random.choice(STATUSES),
        "host":      random.choice(HOSTS),
        "message":   pad,
    }
    out.append('{"index":{}}')
    out.append(json.dumps(doc, separators=(',', ':')))

sys.stdout.write('\n'.join(out) + '\n')
PY
}

# ── Seed function ─────────────────────────────────────────────────────────────

seed_index() {
    local index=$1
    local tmpdir="$WORKDIR/$index"
    mkdir -p "$tmpdir"

    local uploaded=0 active_jobs=0 bid=0

    while [ $uploaded -lt $DOCS_PER_INDEX ]; do
        local remaining=$((DOCS_PER_INDEX - uploaded))
        local batch=$BATCH_SIZE
        [ $remaining -lt $batch ] && batch=$remaining

        local f="$tmpdir/${bid}.ndjson"
        gen_docs "$batch" > "$f"

        (
            $CURL -XPOST "${ES_URL}/${index}/_bulk" \
                -H 'Content-Type: application/x-ndjson' \
                --data-binary "@$f" > /dev/null
            rm -f "$f"
        ) &

        uploaded=$((uploaded + batch))
        bid=$((bid + 1))
        active_jobs=$((active_jobs + 1))
        printf "\r  [${index}] Progress: %d/%d docs (%d%%)" $uploaded $DOCS_PER_INDEX $((uploaded * 100 / DOCS_PER_INDEX))

        if [ $active_jobs -ge $PARALLEL ]; then
            wait
            active_jobs=0
        fi
    done

    wait
}

# ── Main ──────────────────────────────────────────────────────────────────────

echo "Seeding data to ${ES_URL}  (batch=${BATCH_SIZE}, parallel=${PARALLEL})..."

for INDEX in "${INDICES[@]}"; do
    echo ""
    echo "=== Creating index: ${INDEX} ==="

    $CURL -XDELETE "${ES_URL}/${INDEX}" > /dev/null 2>&1 || true

    $CURL -XPUT "${ES_URL}/${INDEX}" -H 'Content-Type: application/json' -d '{
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 0,
            "refresh_interval": "30s"
        }
    }' > /dev/null
    echo "  index created"

    seed_index "$INDEX"

    $CURL -XPOST "${ES_URL}/${INDEX}/_refresh" > /dev/null 2>&1

    SIZE=$($CURL "${ES_URL}/_cat/indices/${INDEX}?h=pri.store.size" 2>/dev/null | tr -d '[:space:]')
    printf "\n  [${INDEX}] Done! Size: %s\n" "$SIZE"
done

echo ""
echo "=== Seed complete ==="
$CURL "${ES_URL}/_cat/indices/demo-*?v&h=health,status,index,pri,rep,docs.count,pri.store.size"
echo ""
