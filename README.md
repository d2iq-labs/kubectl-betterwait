<!--
 Copyright 2023 D2iQ, Inc. All rights reserved.
 SPDX-License-Identifier: Apache-2.0
 -->

# kubectl-betterwait

The motivation for this project is an old outstanding
[kubectl feature request](https://github.com/kubernetes/kubectl/issues/1516).

This can be used either as [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
or as container. The plugin uses the arguments from
[kubectl wait](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_wait/)
to first wait for the specified resources to exist by running `kubectl get` and then runs `kubectl wait`.

## kubectl plugin

```sh
kubectl betterwait --for=condition=established --timeout=1m crds/clusters.cluster.x-k8s.io
```

## container image

```sh
docker run -it                                \
  -v $KUBECONFIG:/kubeconfig                  \
  --env KUBECONFIG=/kubeconfig                \
  ghcr.io/d2iq-labs/kubectl-betterwait:v0.2.0 \
  --for=condition=established --timeout=1m crds/clusters.cluster.x-k8s.io
```
