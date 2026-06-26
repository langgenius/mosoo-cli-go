GO ?= go
BUN ?= bun
MOSOO_REPO ?= https://github.com/langgenius/mosoo.git
MOSOO_REF ?= main
MOSOO_HOST_BASE ?= http://127.0.0.1:8787
LATHE_MODULE := github.com/lathe-cli/lathe/pkg/lathe
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || printf none)
BUILD_DATE ?= $(shell git show -s --format=%cI HEAD 2>/dev/null || printf unknown)
GO_LDFLAGS ?= -X '$(LATHE_MODULE).Version=$(VERSION)' -X '$(LATHE_MODULE).Commit=$(COMMIT)' -X '$(LATHE_MODULE).Date=$(BUILD_DATE)'
EXPECTED_VERSION_OUTPUT := mosoo $(VERSION) ($(COMMIT), $(BUILD_DATE))

GOBIN := $(shell $(GO) env GOBIN)
GOPATH := $(shell $(GO) env GOPATH)
BINDIR ?= $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)
TOOLS_DIR := .cache/bin
LATHE_PKG := github.com/lathe-cli/lathe/cmd/lathe
LATHE_BIN := $(TOOLS_DIR)/lathe
LATHE ?= $(LATHE_BIN)
MOSOO_DIR := .cache/mosoo
SPEC_FILE := docs/openapi/public-thread-api.openapi.json
GRAPHQL_SPEC_FILE := docs/graphql/console.graphql
CONSOLE_REST_SPEC_FILE := docs/openapi/console-rest.openapi.json
SOURCE_NAME := threads
CONSOLE_SOURCE_NAME := console
CONSOLE_REST_SOURCE_NAME := consolerest
PINNED_TAG := local-snapshot
SYNC_DIR := .cache/specs-sync/$(SOURCE_NAME)
CONSOLE_SYNC_DIR := .cache/specs-sync/$(CONSOLE_SOURCE_NAME)
CONSOLE_REST_SYNC_DIR := .cache/specs-sync/$(CONSOLE_REST_SOURCE_NAME)
OVERLAY_DIR := overlays
PUBLISH_SKILL_DIR := publish/skills/mosoo
PUBLISH_CLI_REFERENCE_DIR := $(PUBLISH_SKILL_DIR)/references/cli

.DEFAULT_GOAL := help
.PHONY: help build install verify-install clean tools _codegen

help:
	@printf '%s\n' \
		'make build                  Generate and build bin/mosoo' \
		'make install                Install mosoo to $(BINDIR)/mosoo' \
		'make verify-install        Verify $(BINDIR)/mosoo reports this build metadata' \
		'make tools                  Build local codegen tools under $(TOOLS_DIR)' \
		'make clean                  Remove generated files and caches' \
		'Generated CLI references: $(PUBLISH_CLI_REFERENCE_DIR)' \
		'' \
		'Variables:' \
		'  MOSOO_REPO=$(MOSOO_REPO)' \
		'  MOSOO_REF=$(MOSOO_REF)' \
		'  MOSOO_HOST_BASE=$(MOSOO_HOST_BASE)' \
		'  VERSION=$(VERSION)' \
		'  COMMIT=$(COMMIT)' \
		'  BUILD_DATE=$(BUILD_DATE)' \
		'  BINDIR=$(BINDIR)'

build: _codegen
	cp cli.yaml cmd/mosoo/cli.yaml
	$(GO) build -trimpath -ldflags "$(GO_LDFLAGS)" -o bin/mosoo ./cmd/mosoo

install: build
	mkdir -p "$(BINDIR)"
	install -m 0755 bin/mosoo "$(BINDIR)/mosoo"
	$(MAKE) verify-install

verify-install:
	@test -x "$(BINDIR)/mosoo" || { echo "$(BINDIR)/mosoo is not installed or executable" >&2; exit 1; }
	@test "$$("$(BINDIR)/mosoo" --version)" = "$(EXPECTED_VERSION_OUTPUT)" || { \
		echo "installed mosoo version mismatch" >&2; \
		echo "expected: $(EXPECTED_VERSION_OUTPUT)" >&2; \
		echo "actual:   $$("$(BINDIR)/mosoo" --version)" >&2; \
		exit 1; \
	}

tools: $(LATHE_BIN)

$(LATHE_BIN): go.mod go.sum
	@mkdir -p "$(dir $@)"
	$(GO) build -trimpath -o "$@" "$(LATHE_PKG)"

clean:
	rm -rf .cache bin cmd/mosoo/cli.yaml internal/generated skills "$(PUBLISH_CLI_REFERENCE_DIR)" specs/sources.yaml specs/sources.test.yaml overlays

_codegen: $(LATHE_BIN)
	@mkdir -p .cache specs "$(SYNC_DIR)/docs/openapi" "$(CONSOLE_SYNC_DIR)/docs/graphql" "$(CONSOLE_REST_SYNC_DIR)/docs/openapi"
	@if [ -d "$(MOSOO_DIR)/.git" ]; then \
		git -C "$(MOSOO_DIR)" fetch --all --tags --quiet; \
	else \
		git clone --quiet "$(MOSOO_REPO)" "$(MOSOO_DIR)"; \
	fi
	@if git -C "$(MOSOO_DIR)" rev-parse --verify --quiet "origin/$(MOSOO_REF)" >/dev/null; then \
		git -C "$(MOSOO_DIR)" checkout --quiet -B "$(MOSOO_REF)" "origin/$(MOSOO_REF)"; \
	else \
		git -C "$(MOSOO_DIR)" -c advice.detachedHead=false checkout --quiet "$(MOSOO_REF)"; \
	fi
	git -C "$(MOSOO_DIR)" submodule update --init --recursive
	cd "$(MOSOO_DIR)" && $(BUN) install --frozen-lockfile
	$(BUN) scripts/export-public-api-openapi.ts
	$(BUN) scripts/export-console-graphql.ts
	$(BUN) scripts/export-console-rest-openapi.ts
	MOSOO_HOST_BASE=$(MOSOO_HOST_BASE) $(BUN) scripts/render-sources-yaml.ts
	$(BUN) scripts/render-overlays.ts
	cp "$(MOSOO_DIR)/$(SPEC_FILE)" "$(SYNC_DIR)/$(SPEC_FILE)"
	cp "$(MOSOO_DIR)/$(GRAPHQL_SPEC_FILE)" "$(CONSOLE_SYNC_DIR)/$(GRAPHQL_SPEC_FILE)"
	cp "$(MOSOO_DIR)/$(CONSOLE_REST_SPEC_FILE)" "$(CONSOLE_REST_SYNC_DIR)/$(CONSOLE_REST_SPEC_FILE)"
	"$(LATHE)" codegen -sources specs/sources.yaml -cache .cache -overlay $(OVERLAY_DIR)
	$(BUN) scripts/render-publish-skill.ts
