# Project Setup
PROJECT_NAME := wordpress-operator
PROJECT_REPO := github.com/bitpoke/$(PROJECT_NAME)

PLATFORMS = linux_amd64 darwin_amd64

DOCKER_REGISTRY := docker.io/bitpoke

GO111MODULE=on

include build/makelib/common.mk

IMAGES ?= wordpress-operator

include build/makelib/image.mk

GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/wordpress-operator

include build/makelib/golang.mk
include build/makelib/kubebuilder-v2.mk
include build/makelib/helm.mk

GO_LDFLAGS += -X github.com/bitpoke/wordpress-operator/pkg/version.buildDate=$(BUILD_DATE) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitVersion=$(VERSION) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitCommit=$(GIT_COMMIT) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitTreeState=$(GIT_TREE_STATE)

.PHONY: .kubebuilder.update.chart
.kubebuilder.update.chart: kubebuilder.manifests $(YQ)
	@$(INFO) updating helm RBAC and CRDs from kubebuilder manifests
	@kustomize build config/ > $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml
	@yq w -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml 'metadata.annotations[helm.sh/hook]' crd-install
	@yq w -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml 'metadata.labels[app]' wordpress-operator
	@yq d -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml metadata.creationTimestamp
	@yq d -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml status metadata.creationTimestamp
	@mv $(HELM_CHARTS_DIR)/wordpress-operator/crds/_crds.yaml $(HELM_CHARTS_DIR)/wordpress-operator/crds/crds.yaml

	@cp config/rbac/role.yaml $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml
	@yq m -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml hack/chart-metadata.yaml
	@yq d -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml metadata.creationTimestamp
	@yq w -d'*' -i $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml metadata.name '{{ template "wordpress-operator.fullname" . }}'
	@echo '{{- if .Values.rbac.create }}' > $(HELM_CHARTS_DIR)/wordpress-operator/templates/controller-clusterrole.yaml
	@cat $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/controller-clusterrole.yaml
	@echo '{{- end }}' >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/controller-clusterrole.yaml
	@rm $(HELM_CHARTS_DIR)/wordpress-operator/templates/_rbac.yaml
	@$(OK) updating helm RBAC and CRDs from kubebuilder manifests
.generate.run: .kubebuilder.update.chart
