VERSION := $(shell git describe --tags 2>/dev/null || echo "0.0.0-dev")
REVISION := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_USER := $(shell id -un)@$(shell hostname)
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)

GO_LDFLAGS = \
	-X github.com/prometheus/common/version.Version=$(VERSION) \
	-X github.com/prometheus/common/version.Revision=$(REVISION) \
	-X github.com/prometheus/common/version.Branch=$(BRANCH) \
	-X github.com/prometheus/common/version.BuildUser=$(BUILD_USER) \
	-X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE)

GO_BUILD_FLAGS = -ldflags "$(GO_LDFLAGS)"

GO_ENV = \
	GOOS=linux \
	GOARCH=amd64 \
	GO111MODULE=on

.PHONY: build
build:
	$(GO_ENV) go build $(GO_BUILD_FLAGS)
