# Mutating trace admission controller

[Mutating admission controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) that injects `init trace id` and base64 encoded `span context` into the `trace.kubernetes.io.init` and `trace.kubernetes.io.context` object annotation.

## Quick start

### Kubernetes

First of all, we need a kubernetes(v1.18.5+) cluster, a single-node cluster will make the next job easier.

Use our `kube-apiserver` and `kube-controller-manger` instead of them in cluster:

1. Clone the source code from [here](https://github.com/Hellcatlk/kubernetes/tree/trace-ot).
2. Run `KUBE_BUILD_PLATFORMS=linux/amd64 KUBE_BUILD_CONFORMANCE=n KUBE_BUILD_HYPERKUBE=n make release-images`.
3. Run `docker load -i _output/release-images/amd64/kube-apiserver.tar`.
4. Run `docker load -i _output/release-images/amd64/kube-controller-manager.tar`.
5. Edit `/etc/kubernetes/manifests/kube-apiserver.yaml`, use our `kube-apiserver image` instead of old image.
6. Edit `/etc/kubernetes/manifests/kube-controller-manager.yaml`, use our `kube-controller-manager image` instead of old image.

### Webhook

The included `Makefile` makes these steps straightforward and the available commands are as follows:

- `make build`: build execute file.
- `make docker`: build and load Docker image.
- `make install`: apply certificate configuration and deployment configuration to cluster for the mutating webhook.
- `make remove`: delete resources associated with the mutating webhook from the active cluster.
- `make test-unit`: run unit test.
- `make test-webhook`: test webhook, use deployment replicaset and pod.
- `make deployment`: apply and delete a deployment.
- `make replicaset`: apply and delete a replicaset.
- `make pod`: apply and delete a  pod.
- `make clean`: remove files build by script.
