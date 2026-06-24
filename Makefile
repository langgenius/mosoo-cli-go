GO ?= go
BUN ?= bun
MOSOO_REPO ?= https://github.com/langgenius/mosoo.git
MOSOO_REF ?= main
MOSOO_HOST_BASE ?= http://127.0.0.1:8787

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
PUBLISH_CLI_REFERENCE_FILE := $(PUBLISH_SKILL_DIR)/references/cli.md
PUBLISH_CLI_REFERENCE_DIR := $(PUBLISH_SKILL_DIR)/references/cli
PUBLISH_CLI_REFERENCE_SKILL_ROOT := $(PUBLISH_CLI_REFERENCE_DIR)
PUBLISH_CLI_REFERENCE_SKILL_DIR := $(PUBLISH_CLI_REFERENCE_SKILL_ROOT)/mosoo

.DEFAULT_GOAL := help
.PHONY: help build install clean tools _codegen

help:
	@printf '%s\n' \
		'make build                  Generate and build bin/mosoo' \
		'make install                Install mosoo to $(BINDIR)/mosoo' \
		'make tools                  Build local codegen tools under $(TOOLS_DIR)' \
		'make clean                  Remove generated files and caches' \
		'Generated CLI reference: $(PUBLISH_CLI_REFERENCE_SKILL_DIR)' \
		'' \
		'Variables:' \
		'  MOSOO_REPO=$(MOSOO_REPO)' \
		'  MOSOO_REF=$(MOSOO_REF)' \
		'  MOSOO_HOST_BASE=$(MOSOO_HOST_BASE)' \
		'  BINDIR=$(BINDIR)'

build: _codegen
	cp cli.yaml cmd/mosoo/cli.yaml
	$(GO) build -trimpath -o bin/mosoo ./cmd/mosoo

install: build
	mkdir -p "$(BINDIR)"
	install -m 0755 bin/mosoo "$(BINDIR)/mosoo"

tools: $(LATHE_BIN)

$(LATHE_BIN): go.mod go.sum
	@mkdir -p "$(dir $@)"
	$(GO) build -trimpath -o "$@" "$(LATHE_PKG)"

clean:
	rm -rf .cache bin cmd/mosoo/cli.yaml internal/generated skills "$(PUBLISH_CLI_REFERENCE_FILE)" "$(PUBLISH_CLI_REFERENCE_DIR)" specs/sources.yaml specs/sources.test.yaml overlays

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
	"$(LATHE)" codegen -sources specs/sources.yaml -cache .cache -overlay $(OVERLAY_DIR) -skill-root "$(PUBLISH_CLI_REFERENCE_SKILL_ROOT)"
	$(BUN) scripts/render-publish-skill.ts
