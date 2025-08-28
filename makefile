LOAD_ENV = [ -f .env ] && set -a && source .env && set +a || echo ".env not found, make sure env vars are set"
CMS := cmd/cms/main.go
DISCOVERY := cmd/discovery/main.go
INDEXER := cmd/workers/indexer/main.go
IMPORTER := cmd/workers/importer/main.go

docs/cms/swagger.json: internal/cms/info.go
	swag init -g internal/cms/info.go -o docs/cms --parseDependency --parseInternal --exclude internal/discovery -q

docs/discovery/swagger.json: internal/discovery/info.go
	swag init -g internal/discovery/info.go -o docs/discovery --parseDependency --parseInternal --exclude internal/cms -q

docs: docs/cms/swagger.json docs/discovery/swagger.json
.PHONY: docs

docs-cms:
	swag init -g internal/cms/info.go -o docs/cms --parseDependency --parseInternal --exclude internal/discovery -q

docs-discovery:
	swag init -g internal/discovery/info.go -o docs/discovery --parseDependency --parseInternal --exclude internal/cms -q

dev: docs
	$(LOAD_ENV) && air


bin/cms: $(CMS) docs/cms/swagger.json
	go build -o $@ $<

bin/discovery: $(DISCOVERY) docs/discovery/swagger.json
	go build -o $@ $<

bin/workers/indexer: $(INDEXER)
	go build -o $@ $<

bin/workers/importer: $(IMPORTER)
	go build -o $@ $<

build: bin/cms bin/discovery bin/workers/indexer bin/workers/importer
.PHONY: build

run-cms:
	$(LOAD_ENV) && go run $(CMS)

run-indexer:
	$(LOAD_ENV) && go run $(INDEXER)

run-importer:
	$(LOAD_ENV) && go run $(IMPORTER)

run-discovery:
	$(LOAD_ENV) && go run $(DISCOVERY)

docker-up: docker-services
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-services:
	docker-compose up -d postgres minio redis opensearch opensearch-dashboards jaeger migrate

clean:
	rm -rf bin/
	docker-compose down -v --remove-orphans
	docker system prune -f
.PHONY: clean

lint:
	golangci-lint run
.PHONY: lint

test:
	@go test \
		-shuffle=on \
		-short \
		-timeout=5m \
		./...
.PHONY: test

