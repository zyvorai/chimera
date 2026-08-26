.PHONY: build test vet verify fmt run docker transiva-config

build:
	go build -o bin/chimera ./cmd/chimera

test:
	go test ./...

vet:
	go vet ./...

verify:
	./scripts/verify.sh

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

run:
	go run ./cmd/chimera serve -config config.example.json

docker:
	docker compose up --build

transiva-config:
	./scripts/make-transiva-config.sh > transiva-chimera.yaml
	@echo "wrote transiva-chimera.yaml"
