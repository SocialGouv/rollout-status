REPO = clusterise/rollout-status
TAG = 1.2

all: build publish

.PHONY: build
build: Dockerfile preflight
	docker build -t $(REPO):$(TAG) .

.PHONY: publish
publish: build
	docker push $(REPO):$(TAG)

.PHONY: preflight
preflight:
	go mod vendor
	go fmt github.com/socialgouv/rollout-status/...

.PHONY: test
test:
	go test github.com/socialgouv/rollout-status/...

.PHONY: cleanup
cleanup:
	rm -rf vendor
