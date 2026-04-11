.PHONY: dev-init dev-start dev-stop dev-destroy build test run seed-data generate-indices

dev-init:
	docker compose up -d
	@echo "Waiting for Elasticsearch to be healthy..."
	@until curl -s -u elastic:elastic http://localhost:9200/_cluster/health | grep -q '"status"'; do sleep 2; done
	@curl -s -u elastic:elastic -XPOST "http://localhost:9200/_security/user/kibana_system/_password" \
		-H 'Content-Type: application/json' -d '{"password":"elastic"}' > /dev/null
	@echo "Kibana system user configured. Kibana at http://localhost:5601"

dev-start:
	docker compose start

dev-stop:
	docker compose stop

dev-destroy:
	docker compose down -v

build:
	go build -o es-cli ./cmd/es-cli

test:
	go test ./... -v -race -cover

run: build
	./es-cli

seed-data:
	go run ./cmd/seed

NUM ?= 100
generate-indices:
	./scripts/generate-indices.sh $(NUM)
