# Copyright 2024 D2iQ, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

GORELEASER_PARALLELISM ?= $(shell nproc --ignore=1)
GORELEASER_VERBOSE ?= false

ifndef GORELEASER_CURRENT_TAG
export GORELEASER_CURRENT_TAG=$(GIT_TAG)
endif

.PHONY: build-snapshot
build-snapshot: ## Builds a snapshot with goreleaser
build-snapshot: dockerauth ; $(info $(M) building snapshot $*)
	goreleaser --verbose=$(GORELEASER_VERBOSE) \
		build \
		--snapshot \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		$(if $(BUILD_ALL),,--single-target)

.PHONY: release
release: ## Builds a release with goreleaser
release: dockerauth ; $(info $(M) building release $*)
	goreleaser --verbose=$(GORELEASER_VERBOSE) \
		release \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		--timeout=60m \
		$(GORELEASER_FLAGS)

.PHONY: release-snapshot
release-snapshot: ## Builds a snapshot release with goreleaser
release-snapshot: dockerauth ; $(info $(M) building snapshot release $*)
	goreleaser --verbose=$(GORELEASER_VERBOSE) \
		release \
		--snapshot \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		--timeout=60m
