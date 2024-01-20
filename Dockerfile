# syntax=docker/dockerfile:1.4

# Copyright 2024 D2iQ, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG KUBECTL_VERSION
FROM --platform=linux/amd64 registry.k8s.io/kubectl-amd64:${KUBECTL_VERSION} as linux-amd64
FROM --platform=linux/arm64 registry.k8s.io/kubectl-arm64:${KUBECTL_VERSION} as linux-arm64

# hadolint ignore=DL3006,DL3029
FROM --platform=linux/${TARGETARCH} linux-${TARGETARCH}

COPY kubectl-betterwait /bin/kubectl-betterwait

# set an environment variable with the path to kubectl in the base image
ENV KUBECTL_EXECUTABLE=/bin/kubectl

# Use uid of nonroot user (65532) because kubernetes expects numeric user when applying pod security policies
USER 65532
ENTRYPOINT ["/bin/kubectl-betterwait"]
