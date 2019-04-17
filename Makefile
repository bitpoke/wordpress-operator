# Image URL to use all building/pushing image targets
APP_VERSION ?= $(shell git describe --abbrev=5 --dirty --tags --always)
REGISTRY := quay.io/presslabs
IMAGE_NAME := wordpress-operator
BUILD_TAG := build
IMAGE_TAGS := $(APP_VERSION)
KUBEBUILDER_VERSION ?= 1.0.4
BINDIR ?= $(PWD)/bin
BUILDDIR ?= $(PWD)/build
CHARTDIR ?= $(PWD)/chart/wordpress-operator

GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH ?= amd64

PATH := $(BINDIR):$(PATH)
SHELL := env PATH=$(PATH) /bin/sh

all: test manager

# Run tests
test: generate manifests
	KUBEBUILDER_ASSETS=$(BINDIR) ginkgo \
		--randomizeAllSpecs --randomizeSuites --failOnPending \
		--cover --coverprofile cover.out --trace --race \
		./pkg/... ./cmd/...

# Build manager binary
manager: generate
	go build -o bin/manager github.com/presslabs/wordpress-operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
	# CRDs
	awk 'FNR==1 && NR!=1 {print "---"}{print}' config/crds/*.yaml > $(CHARTDIR)/templates/_crds.yaml
	yq m -d'*' -i $(CHARTDIR)/templates/_crds.yaml hack/chart-metadata.yaml
	yq w -d'*' -i $(CHARTDIR)/templates/_crds.yaml 'metadata.annotations[helm.sh/hook]' crd-install
	yq d -d'*' -i $(CHARTDIR)/templates/_crds.yaml metadata.creationTimestamp
	yq d -d'*' -i $(CHARTDIR)/templates/_crds.yaml status metadata.creationTimestamp
	# add shortName to CRD until https://github.com/kubernetes-sigs/kubebuilder/issues/404 is solved
	yq w -i $(CHARTDIR)/templates/_crds.yaml 'spec.names.shortNames[0]' wp
	echo '{{- if and .Values.crd.install (not (.Capabilities.APIVersions.Has "wordpress.presslabs.org/v1alpha1")) }}' > $(CHARTDIR)/templates/crds.yaml
	cat $(CHARTDIR)/templates/_crds.yaml >> $(CHARTDIR)/templates/crds.yaml
	echo '{{- end }}' >> $(CHARTDIR)/templates/crds.yaml
	rm $(CHARTDIR)/templates/_crds.yaml
	# RBAC
	cp config/rbac/rbac_role.yaml $(CHARTDIR)/templates/_rbac.yaml
	yq m -d'*' -i $(CHARTDIR)/templates/_rbac.yaml hack/chart-metadata.yaml
	yq d -d'*' -i $(CHARTDIR)/templates/_rbac.yaml metadata.creationTimestamp
	yq w -d'*' -i $(CHARTDIR)/templates/_rbac.yaml metadata.name '{{ template "wordpress-operator.fullname" . }}'
	echo '{{- if .Values.rbac.create }}' > $(CHARTDIR)/templates/controller-clusterrole.yaml
	cat $(CHARTDIR)/templates/_rbac.yaml >> $(CHARTDIR)/templates/controller-clusterrole.yaml
	echo '{{- end }}' >> $(CHARTDIR)/templates/controller-clusterrole.yaml
	rm $(CHARTDIR)/templates/_rbac.yaml

.PHONY: chart
chart:
	yq w -i $(CHARTDIR)/Chart.yaml version "$(APP_VERSION)"
	yq w -i $(CHARTDIR)/Chart.yaml appVersion "$(APP_VERSION)"
	mv $(CHARTDIR)/values.yaml $(CHARTDIR)/_values.yaml
	sed 's#$(REGISTRY)/$(IMAGE_NAME):latest#$(REGISTRY)/$(IMAGE_NAME):$(APP_VERSION)#g' $(CHARTDIR)/_values.yaml > $(CHARTDIR)/values.yaml
	rm $(CHARTDIR)/_values.yaml

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
.PHONY: images
images:
	docker build . -t $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG)
	set -e; \
		for tag in $(IMAGE_TAGS); do \
			docker tag $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
	done

# Push the docker image
.PHONY: publish
publish: images
	set -e; \
		for tag in $(IMAGE_TAGS); do \
		docker push $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
	done

lint:
	$(BINDIR)/golangci-lint run ./pkg/... ./cmd/...
	hack/license-check

dependencies:
	test -d $(BINDIR) || mkdir $(BINDIR)
	GOBIN=$(BINDIR) go install ./vendor/github.com/onsi/ginkgo/ginkgo
	curl -sL https://github.com/mikefarah/yq/releases/download/2.1.1/yq_$(GOOS)_$(GOARCH) -o $(BINDIR)/yq
	chmod +x $(BINDIR)/yq
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $(BINDIR) v1.12.2
	curl -sL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH).tar.gz | \
		tar -zx -C $(BINDIR) --strip-components=2
