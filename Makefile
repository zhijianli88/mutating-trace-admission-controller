IMAGE=trace-context-injector:v1
OUTPUT=bin

.PHONY: build
build:
	@mkdir -p $(OUTPUT)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o $(OUTPUT)/webhook cmd/webhook/main.go

docker: build
	@docker rmi -f $(IMAGE)
	@docker build -f build/Dockerfile -t $(IMAGE) .
	@docker save -o $(OUTPUT)/$(IMAGE).tar $(IMAGE)

install:
	@hack/webhook-create-signed-cert.sh --service trace-context-injector-webhook-svc --secret trace-context-injector-webhook-certs --namespace default
	@cat deploy/base/mutatingwebhook.yaml | hack/webhook-patch-ca-bundle.sh > deploy/base/mutatingwebhook-ca-bundle.yaml
	@./tools/kustomize build deploy/base | kubectl apply -f -

remove:
	@kubectl delete secret trace-context-injector-webhook-certs
	@./tools/kustomize build deploy/base | kubectl delete -f -

test-unit:
	@go test ./... -coverprofile=cover.out
	@go tool cover -html=cover.out -o coverage.html

test-webhook: deployment replicaset pod

deployment:
	@kubectl apply -f test/yaml/deployment.yaml
	@sleep 1
	@kubectl apply -f test/yaml/deployment_v2.yaml
	@sleep 1
	@kubectl delete -f test/yaml/deployment_v2.yaml

replicaset:
	@kubectl apply -f test/yaml/replicaset.yaml
	@sleep 1
	@kubectl apply -f test/yaml/replicaset_v2.yaml
	@sleep 1
	@kubectl delete -f test/yaml/replicaset_v2.yaml

pod:
	@kubectl apply -f test/yaml/pod.yaml
	@sleep 1
	@kubectl apply -f test/yaml/pod_v2.yaml
	@sleep 1
	@kubectl delete -f test/yaml/pod_v2.yaml

clean:
	@rm -f $(OUTPUT)/*
	@rm -f deploy/base/mutatingwebhook-ca-bundle.yaml
	@docker rmi -f $(IMAGE)
