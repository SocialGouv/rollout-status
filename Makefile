all: build

.PHONY: build
build: preflight
	CGO_ENABLED=0 GOOS=linux go build -a -mod vendor -installsuffix cgo -o rollout-status github.com/socialgouv/rollout-status/cmd

.PHONY: preflight
preflight:
	go mod vendor
	go fmt github.com/socialgouv/rollout-status/...

.PHONY: test
test:
	go test github.com/socialgouv/rollout-status/...
