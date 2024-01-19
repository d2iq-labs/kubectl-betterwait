# syntax=docker/dockerfile:1.4

# Copyright 2024 D2iQ, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# FIXME set this based on the .Version
FROM --platform=linux/amd64 registry.k8s.io/kubectl-amd64:v1.28.6 as linux-amd64
FROM --platform=linux/arm64 registry.k8s.io/kubectl-arm64:v1.28.6 as linux-arm64

# hadolint ignore=DL3006,DL3029
FROM --platform=linux/${TARGETARCH} linux-${TARGETARCH}

COPY kubectl-betterwait /bin/kubectl-betterwait

# Use uid of nonroot user (65532) because kubernetes expects numeric user when applying pod security policies
USER 65532
ENTRYPOINT ["/bin/kubectl-betterwait"]