# Project Setup
PROJECT_NAME := wordpress-operator
PROJECT_REPO := github.com/bitpoke/$(PROJECT_NAME)

PLATFORMS = linux_amd64 darwin_amd64

include build/makelib/common.mk

GO111MODULE=on
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/wordpress-operator
GOLANGCI_LINT_VERSION = 1.37.1
GO_LDFLAGS += -X github.com/bitpoke/wordpress-operator/pkg/version.buildDate=$(BUILD_DATE) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitVersion=$(VERSION) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitCommit=$(GIT_COMMIT) \
	       -X github.com/bitpoke/wordpress-operator/pkg/version.gitTreeState=$(GIT_TREE_STATE)
include build/makelib/golang.mk

DOCKER_REGISTRY := docker.io/bitpoke
IMAGES ?= wordpress-operator
include build/makelib/image.mk

GEN_CRD_OPTIONS := crd:crdVersions=v1,preserveUnknownFields=false
include build/makelib/kubebuilder-v3.mk

# fix for https://github.com/kubernetes-sigs/controller-tools/issues/476
.PHONY: .kubebuilder.fix-preserve-unknown-fields
.kubebuilder.fix-preserve-unknown-fields:
		@for crd in $(wildcard $(CRD_DIR)/*.yaml) ; do \
			$(YQ) e '.spec.preserveUnknownFields=false' -i "$${crd}" ;\
		done
.kubebuilder.manifests.done: .kubebuilder.fix-preserve-unknown-fields

include build/makelib/helm.mk

.PHONY: .kubebuilder.update.chart
.kubebuilder.update.chart: kubebuilder.manifests $(YQ)
	@$(INFO) updating helm RBAC and CRDs from kubebuilder manifests
	@rm -rf $(HELM_CHARTS_DIR)/wordpress-operator/crds
	@mkdir -p $(HELM_CHARTS_DIR)/wordpress-operator/crds
	@set -e; \
		for crd in $(wildcard $(CRD_DIR)/*.yaml) ; do \
			cp $${crd} $(HELM_CHARTS_DIR)/wordpress-operator/crds/ ; \
			$(YQ) e '.metadata.labels["app"]="wordpress-operator"'        -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/$$(basename $${crd}) ; \
			$(YQ) e 'del(.metadata.creationTimestamp)'                    -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/$$(basename $${crd}) ; \
			$(YQ) e 'del(.status)'                                        -i $(HELM_CHARTS_DIR)/wordpress-operator/crds/$$(basename $${crd}) ; \
		done
	@echo '{{- if .Values.rbac.create }}'                                 > $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo 'apiVersion: rbac.authorization.k8s.io/v1'                     >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo 'kind: ClusterRole'                                            >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo 'metadata:'                                                    >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo '  name: {{ include "wordpress-operator.fullname" . }}'        >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo '  labels:'                                                    >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo '    {{- include "wordpress-operator.labels" . | nindent 4 }}' >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo 'rules:'                                                       >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@yq e -P '.rules' config/rbac/role.yaml                              >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@echo '{{- end }}'                                                   >> $(HELM_CHARTS_DIR)/wordpress-operator/templates/clusterrole.yaml
	@$(OK) updating helm RBAC and CRDs from kubebuilder manifests
.generate.run: .kubebuilder.update.chart

.PHONY: .helm.publish
.helm.publish:
	@$(INFO) publishing helm charts
	@rm -rf $(WORK_DIR)/charts
	@git clone -q git@github.com:bitpoke/helm-charts.git $(WORK_DIR)/charts
	@cp $(HELM_OUTPUT_DIR)/*.tgz $(WORK_DIR)/charts/docs/
	@git -C $(WORK_DIR)/charts add $(WORK_DIR)/charts/docs/*.tgz
	@git -C $(WORK_DIR)/charts commit -q -m "Added $(call list-join,$(COMMA)$(SPACE),$(foreach c,$(HELM_CHARTS),$(c)-v$(HELM_CHART_VERSION)))"
	@git -C $(WORK_DIR)/charts push -q
	@$(OK) publishing helm charts
.publish.run: .helm.publish

.PHONY: .helm.package.prepare.wordpress-operator
.helm.package.prepare.wordpress-operator:  $(YQ)
	@$(INFO) prepare wordpress-operator chart $(HELM_CHART_VERSION)
	@$(YQ) e '.image="$(DOCKER_REGISTRY)/wordpress-operator:$(IMAGE_TAG)"' -i $(HELM_CHARTS_WORK_DIR)/wordpress-operator/values.yaml
	@$(SED) 's/:latest/:$(IMAGE_TAG)/g' $(HELM_CHARTS_WORK_DIR)/wordpress-operator/Chart.yaml
	@$(OK) prepare wordpress-operator chart $(HELM_CHART_VERSION)
.helm.package.run.wordpress-operator: .helm.package.prepare.wordpress-operator

