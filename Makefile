IMAGE=trace-context-injector:v1

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o bin/webhook cmd/webhook/main.go

docker:
	docker build -f build/Dockerfile -t $(IMAGE) .

install:
	hack/webhook-create-signed-cert.sh --service trace-context-injector-webhook-svc --secret trace-context-injector-webhook-certs --namespace default
	cat deploy/base/mutatingwebhook.yaml | hack/webhook-patch-ca-bundle.sh > deploy/base/mutatingwebhook-ca-bundle.yaml
	kustomize build deploy/base | kubectl apply -f -

remove:
	kubectl delete secret trace-context-injector-webhook-certs
	kustomize build deploy/base | kubectl delete -f -

.PHONY: test
test:
	@kubectl apply -f test/yaml/Deployment.yaml
	@kubectl delete -f test/yaml/Deployment.yaml

clean:
	rm deploy/base/mutatingwebhook-ca-bundle.yaml
	docker rmi -f $(IMAGE)
