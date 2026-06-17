GO ?= go
BUN ?= bun
LATHE ?= lathe
MOSOO_REPO ?= https://github.com/langgenius/mosoo.git
MOSOO_REF ?= main
INSTALL_DIR ?= $(HOME)/.bin

MOSOO_DIR := .cache/mosoo
SPEC_FILE := docs/openapi/public-thread-api.openapi.json
SOURCE_NAME := threads
PINNED_TAG := local-snapshot

.PHONY: all mosoo openapi sources sync-cache codegen embed-manifest build install clean

all: build

mosoo:
	@mkdir -p .cache
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

openapi: mosoo
	cd "$(MOSOO_DIR)" && $(BUN) install --frozen-lockfile
	$(BUN) scripts/export-public-api-openapi.ts

sources: openapi
	@mkdir -p specs
	@printf '%s\n' \
		'sources:' \
		'  $(SOURCE_NAME):' \
		'    display_name: public-thread-api' \
		'    repo_url: file://$(CURDIR)/$(MOSOO_DIR)' \
		'    pinned_tag: $(PINNED_TAG)' \
		'    backend: openapi3' \
		'    openapi3:' \
		'      files:' \
		'        - $(SPEC_FILE)' \
		> specs/sources.yaml

sync-cache: sources
	@mkdir -p ".cache/specs-sync/$(SOURCE_NAME)/docs/openapi"
	cp "$(MOSOO_DIR)/$(SPEC_FILE)" ".cache/specs-sync/$(SOURCE_NAME)/$(SPEC_FILE)"
	@printf '%s\n' \
		'source: $(SOURCE_NAME)' \
		'backend: openapi3' \
		'synced_from: $(PINNED_TAG)' \
		'resolved_sha: $(PINNED_TAG)' \
		> ".cache/specs-sync/$(SOURCE_NAME)/sync-state.yaml"

codegen: sync-cache
	$(LATHE) codegen -sources specs/sources.yaml -cache .cache

embed-manifest:
	cp cli.yaml cmd/mosoo/cli.yaml

build: codegen embed-manifest
	$(GO) build -trimpath -o bin/mosoo ./cmd/mosoo

install: build
	mkdir -p "$(INSTALL_DIR)"
	install -m 0755 bin/mosoo "$(INSTALL_DIR)/mosoo"

clean:
	rm -rf .cache bin cmd/mosoo/cli.yaml internal/generated skills specs/sources.yaml
