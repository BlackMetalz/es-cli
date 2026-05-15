#!/bin/bash
# seed-data-alt.sh — thematic demo data: ecom-orders, app-logs, iot-sensors
# Uses Python3 for fast bulk generation + parallel curl uploads
#
# Usage:
#   ./seed-data-alt.sh
#   ES_URL=http://remote:9200 PARALLEL=8 BATCH_SIZE=20000 ./seed-data-alt.sh

ES_URL="${ES_URL:-http://localhost:9200}"
ES_USER="${ES_USER:-elastic}"
ES_PASS="${ES_PASS:-elastic}"
CURL="curl -s -u ${ES_USER}:${ES_PASS}"

BATCH_SIZE="${BATCH_SIZE:-10000}"   # docs per bulk request
PARALLEL="${PARALLEL:-4}"           # concurrent bulk requests

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

# ── Mappings ──────────────────────────────────────────────────────────────────

MAPPING_ECOM=$(cat <<'JSON'
{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0, "refresh_interval": "30s"},
  "mappings": {"properties": {
    "@timestamp":     {"type": "date"},
    "order_id":       {"type": "keyword"},
    "customer_id":    {"type": "keyword"},
    "status":         {"type": "keyword"},
    "total_usd":      {"type": "float"},
    "items_count":    {"type": "integer"},
    "payment_method": {"type": "keyword"},
    "country":        {"type": "keyword"},
    "channel":        {"type": "keyword"},
    "discount_pct":   {"type": "integer"},
    "shipping_days":  {"type": "integer"},
    "notes":          {"type": "text", "index": false}
  }}
}
JSON
)

MAPPING_LOGS=$(cat <<'JSON'
{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0, "refresh_interval": "30s"},
  "mappings": {"properties": {
    "@timestamp":   {"type": "date"},
    "level":        {"type": "keyword"},
    "service":      {"type": "keyword"},
    "host":         {"type": "keyword"},
    "trace_id":     {"type": "keyword"},
    "span_id":      {"type": "keyword"},
    "status_code":  {"type": "integer"},
    "duration_ms":  {"type": "integer"},
    "method":       {"type": "keyword"},
    "path":         {"type": "keyword"},
    "message":      {"type": "text", "index": false},
    "error_class":  {"type": "keyword"}
  }}
}
JSON
)

MAPPING_IOT=$(cat <<'JSON'
{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0, "refresh_interval": "30s"},
  "mappings": {"properties": {
    "@timestamp":   {"type": "date"},
    "device_id":    {"type": "keyword"},
    "device_type":  {"type": "keyword"},
    "facility":     {"type": "keyword"},
    "zone":         {"type": "keyword"},
    "reading":      {"type": "float"},
    "unit":         {"type": "keyword"},
    "battery_pct":  {"type": "integer"},
    "signal_rssi":  {"type": "integer"},
    "alert":        {"type": "boolean"},
    "alert_reason": {"type": "keyword"},
    "firmware_ver": {"type": "keyword"},
    "raw_payload":  {"type": "keyword", "index": false}
  }}
}
JSON
)

# ── Python generators ─────────────────────────────────────────────────────────

gen_ecom_orders() {
    python3 - "$1" <<'PY'
import sys, random, string, datetime, json

count = int(sys.argv[1])

STATUSES  = ['pending', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded']
PAYMENTS  = ['credit_card', 'paypal', 'apple_pay', 'crypto', 'bank_transfer']
COUNTRIES = ['US', 'GB', 'DE', 'FR', 'JP', 'AU', 'CA', 'BR', 'IN', 'SG', 'MX', 'NL', 'SE', 'PL', 'ZA']
CHANNELS  = ['web', 'mobile_ios', 'mobile_android', 'marketplace', 'phone']
NOTES     = [
    'standard shipment', 'gift wrap requested', 'express delivery',
    'fragile - handle with care', 'leave at door', 'signature required',
    'po box delivery', 'contactless drop-off', 'customer requested delay',
    'bulk order - verify inventory before dispatch', 'priority handling required',
    'temperature-sensitive item', 'oversized package - freight carrier',
]

now      = int(datetime.datetime.utcnow().timestamp())
RANGE_S  = 90 * 86400
HEX      = '0123456789abcdef'

out = []
for _ in range(count):
    ts  = datetime.datetime.utcfromtimestamp(now - random.randint(0, RANGE_S))
    dis = 0 if random.random() < 0.6 else random.randint(5, 50)
    doc = {
        "@timestamp":     ts.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
        "order_id":       "ord-" + "".join(random.choices(HEX, k=8)),
        "customer_id":    f"cust-{random.randint(1, 5000)}",
        "status":         random.choice(STATUSES),
        "total_usd":      round(random.uniform(5.0, 2500.0), 2),
        "items_count":    random.randint(1, 20),
        "payment_method": random.choice(PAYMENTS),
        "country":        random.choice(COUNTRIES),
        "channel":        random.choice(CHANNELS),
        "discount_pct":   dis,
        "shipping_days":  random.randint(1, 14),
        "notes":          random.choice(NOTES),
    }
    out.append('{"index":{}}')
    out.append(json.dumps(doc, separators=(',', ':')))

sys.stdout.write("\n".join(out) + "\n")
PY
}

gen_app_logs() {
    python3 - "$1" <<'PY'
import sys, random, datetime, json

count = int(sys.argv[1])

# weighted: DEBUG 10%, INFO 55%, WARN 20%, ERROR 13%, FATAL 2%
LEVELS   = ['DEBUG']*10 + ['INFO']*55 + ['WARN']*20 + ['ERROR']*13 + ['FATAL']*2
SERVICES = ['api-gateway', 'auth-svc', 'order-svc', 'inventory-svc', 'payment-svc', 'notification-svc']
METHODS  = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']
PATHS    = [
    '/api/v1/orders', '/api/v1/users', '/api/v1/products',
    '/api/v1/auth/login', '/api/v1/auth/logout', '/api/v1/payments',
    '/api/v1/inventory', '/api/v1/notifications', '/api/v1/reports', '/health',
]
# weighted toward 2xx
STATUS_CODES = [200]*50 + [201]*10 + [204]*5 + [400]*8 + [401]*5 + \
               [403]*3 + [404]*8 + [429]*3 + [500]*5 + [502]*2 + [503]*1
ERRORS = [
    'NullPointerException', 'TimeoutError', 'ConnectionRefused',
    'OutOfMemoryError', 'DatabaseError', 'AuthenticationFailed',
    'RateLimitExceeded', 'CircuitBreakerOpen', 'ValidationError',
]
MSGS = {
    'DEBUG': ['cache miss for key {k}', 'executing query plan {k}', 'loaded config revision {k}', 'gc pause {k}ms'],
    'INFO':  ['request processed in {k}ms', 'user {k} authenticated', 'order {k} created',
              'payment {k} authorized', 'cache warmed {k} entries', 'job {k} completed'],
    'WARN':  ['slow query {k}ms exceeds threshold', 'retry {k} for upstream call', 'memory at {k}%'],
    'ERROR': ['request failed after {k} retries', 'unhandled exception in handler {k}', 'upstream timeout {k}ms'],
    'FATAL': ['process {k} crashed with signal 11', 'critical data loss detected in partition {k}'],
}
HEX     = '0123456789abcdef'
now     = int(datetime.datetime.utcnow().timestamp())
RANGE_S = 7 * 86400

out = []
for _ in range(count):
    ts    = datetime.datetime.utcfromtimestamp(now - random.randint(0, RANGE_S))
    level = random.choice(LEVELS)
    # log-normal-ish duration: 80% fast, 20% slow
    dur   = random.randint(1, 500) if random.random() < 0.8 else random.randint(500, 30000)
    k     = random.randint(1, 9999)
    msg   = random.choice(MSGS[level]).format(k=k)
    doc = {
        "@timestamp":  ts.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
        "level":       level,
        "service":     random.choice(SERVICES),
        "host":        f"ip-10-0-{random.randint(1,4)}-{random.randint(1,20)}",
        "trace_id":    "".join(random.choices(HEX, k=16)),
        "span_id":     "".join(random.choices(HEX, k=8)),
        "status_code": random.choice(STATUS_CODES),
        "duration_ms": dur,
        "method":      random.choice(METHODS),
        "path":        random.choice(PATHS),
        "message":     msg,
    }
    if level in ('ERROR', 'FATAL'):
        doc['error_class'] = random.choice(ERRORS)
    out.append('{"index":{}}')
    out.append(json.dumps(doc, separators=(',', ':')))

sys.stdout.write("\n".join(out) + "\n")
PY
}

gen_iot_sensors() {
    python3 - "$1" <<'PY'
import sys, random, datetime, json

count = int(sys.argv[1])

TYPES     = ['temperature', 'humidity', 'pressure', 'vibration', 'power_meter', 'flow_meter']
RANGES    = {
    'temperature': (-20.0, 150.0),
    'humidity':    (0.0, 100.0),
    'pressure':    (900.0, 1100.0),
    'vibration':   (0.0, 50.0),
    'power_meter': (0.0, 500.0),
    'flow_meter':  (0.0, 1000.0),
}
UNITS     = {
    'temperature': 'celsius', 'humidity': 'percent', 'pressure': 'hPa',
    'vibration': 'mm/s', 'power_meter': 'kW', 'flow_meter': 'L/min',
}
WIRED        = {'power_meter', 'flow_meter'}
FACILITIES   = ['plant-a', 'plant-b', 'plant-c', 'plant-d', 'warehouse-1', 'warehouse-2']
FIRMWARES    = ['v2.1.0', 'v2.2.0', 'v2.3.1', 'v3.0.0']
ALERT_REASONS= ['threshold_exceeded', 'sensor_fault', 'battery_low', 'offline']

# Build a reusable hex pool to avoid per-doc random generation for raw_payload
HEX  = '0123456789abcdef'
POOL = "".join(random.choices(HEX, k=8192))

now     = int(datetime.datetime.utcnow().timestamp())
RANGE_S = 30 * 86400

out = []
for i in range(count):
    ts    = datetime.datetime.utcfromtimestamp(now - random.randint(0, RANGE_S))
    dtype = random.choice(TYPES)
    lo, hi= RANGES[dtype]
    alert = random.random() < 0.05
    # reuse pool with rotating offset — fast, no extra allocation
    off   = (i * 300) % (len(POOL) - 320)
    raw   = POOL[off: off + 300 + (i % 20)]

    doc = {
        "@timestamp":   ts.strftime("%Y-%m-%dT%H:%M:%S.000Z"),
        "device_id":    f"sensor-{random.randint(1, 500)}",
        "device_type":  dtype,
        "facility":     random.choice(FACILITIES),
        "zone":         f"zone-{random.randint(1, 10)}",
        "reading":      round(random.uniform(lo, hi), 2),
        "unit":         UNITS[dtype],
        "signal_rssi":  random.randint(-100, -40),
        "alert":        alert,
        "firmware_ver": random.choice(FIRMWARES),
        "raw_payload":  raw,
    }
    if dtype not in WIRED:
        doc['battery_pct'] = random.randint(0, 100)
    if alert:
        reason = random.choice(ALERT_REASONS)
        if dtype in WIRED and reason == 'battery_low':
            reason = 'threshold_exceeded'
        doc['alert_reason'] = reason

    out.append('{"index":{}}')
    out.append(json.dumps(doc, separators=(',', ':')))

sys.stdout.write("\n".join(out) + "\n")
PY
}

# ── Seed function ─────────────────────────────────────────────────────────────

seed_index() {
    local index=$1
    local total=$2
    local generator=$3
    local tmpdir="$WORKDIR/$index"
    mkdir -p "$tmpdir"

    local uploaded=0 active_jobs=0 bid=0

    while [ $uploaded -lt $total ]; do
        local remaining=$((total - uploaded))
        local batch=$BATCH_SIZE
        [ $remaining -lt $batch ] && batch=$remaining

        local f="$tmpdir/${bid}.ndjson"

        # Generate data synchronously (Python is fast — no per-doc subprocesses)
        $generator "$batch" > "$f"

        # Upload asynchronously
        (
            $CURL -XPOST "${ES_URL}/${index}/_bulk" \
                -H 'Content-Type: application/x-ndjson' \
                --data-binary "@$f" > /dev/null
            rm -f "$f"
        ) &

        uploaded=$((uploaded + batch))
        bid=$((bid + 1))
        active_jobs=$((active_jobs + 1))
        printf "\r  [%-13s] %d/%d docs (%d%%)" "$index" $uploaded $total $((uploaded * 100 / total))

        if [ $active_jobs -ge $PARALLEL ]; then
            wait
            active_jobs=0
        fi
    done

    wait  # flush remaining uploads
}

# ── Main ──────────────────────────────────────────────────────────────────────

echo "Seeding to ${ES_URL}  (batch=${BATCH_SIZE}, parallel=${PARALLEL})"

for INDEX in "ecom-orders" "app-logs" "iot-sensors"; do
    echo ""
    echo "=== ${INDEX} ==="

    # Pick docs count and mapping per index
    case $INDEX in
        ecom-orders) DOCS=51200; MAPPING="$MAPPING_ECOM"; GENERATOR="gen_ecom_orders" ;;
        app-logs)    DOCS=51200; MAPPING="$MAPPING_LOGS"; GENERATOR="gen_app_logs"    ;;
        iot-sensors) DOCS=51200; MAPPING="$MAPPING_IOT";  GENERATOR="gen_iot_sensors" ;;
    esac

    $CURL -XDELETE "${ES_URL}/${INDEX}" > /dev/null 2>&1 || true

    $CURL -XPUT "${ES_URL}/${INDEX}" \
        -H 'Content-Type: application/json' \
        -d "$MAPPING" > /dev/null
    echo "  mapping created"

    seed_index "$INDEX" "$DOCS" "$GENERATOR"

    $CURL -XPOST "${ES_URL}/${INDEX}/_refresh" > /dev/null 2>&1

    SIZE=$($CURL "${ES_URL}/_cat/indices/${INDEX}?h=pri.store.size" 2>/dev/null | tr -d '[:space:]')
    printf "\n  done — size: %s\n" "$SIZE"
done

echo ""
echo "=== Summary ==="
$CURL "${ES_URL}/_cat/indices/ecom-orders,app-logs,iot-sensors?v&h=health,status,index,docs.count,pri.store.size"
echo ""
