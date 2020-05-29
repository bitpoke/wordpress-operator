# Copyright 2019 Pressinfra Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ifndef __KUBEBUILDERV2_MAKEFILE__
__KUBEBUILDERV2_MAKEFILE__ := included

# ====================================================================================
# Options

KUBEBUILDER_VERSION ?= 2.2.0
KUBEBUILDER := $(TOOLS_HOST_DIR)/kubebuilder-$(KUBEBUILDER_VERSION)

CRD_DIR ?= config/crds
API_DIR ?= pkg/apis

CRD_OPTIONS ?= "crd:trivialVersions=true"
BOILERPLATE_FILE ?= ./hack/boilerplate.go.txt

# these are use by the kubebuilder test harness

TEST_ASSET_KUBE_APISERVER := $(KUBEBUILDER)/kube-apiserver
TEST_ASSET_ETCD := $(KUBEBUILDER)/etcd
TEST_ASSET_KUBECTL := $(KUBEBUILDER)/kubectl
export TEST_ASSET_KUBE_APISERVER TEST_ASSET_ETCD TEST_ASSET_KUBECTL

# ====================================================================================
# Setup environment

include $(COMMON_SELF_DIR)/golang.mk

# ====================================================================================
# tools

# kubebuilder download and install
$(KUBEBUILDER):
	@echo ${TIME} ${BLUE}[TOOL]${CNone} installing kubebuilder $(KUBEBUILDER_VERSION)
	@mkdir -p $(TOOLS_HOST_DIR)/tmp || $(FAIL)
	@curl -fsSL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOHOSTOS)_$(GOHOSTARCH).tar.gz | tar -xz -C $(TOOLS_HOST_DIR)/tmp  || $(FAIL)
	@mv $(TOOLS_HOST_DIR)/tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOHOSTOS)_$(GOHOSTARCH)/bin $(KUBEBUILDER) || $(FAIL)
	@rm -fr $(TOOLS_HOST_DIR)/tmp
	@$(OK) installing kubebuilder $(KUBEBUILDER_VERSION)

$(eval $(call tool.go.get,controller-gen,v0.2.4,sigs.k8s.io/controller-tools/cmd/controller-gen))

# ====================================================================================
# Kubebuilder Targets

$(eval $(call common.target,kubebuilder.manifests))

# Generate manifests e.g. CRD, RBAC etc.
.do.kubebuilder.manifests: $(CONTROLLER_GEN)
	@$(INFO) Generating Kubebuilder manifests
	@# first delete the CRD_DIR, to remove the CRDs of types that no longer exist
	@rm -rf $(CRD_DIR)

	@$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=$(CRD_DIR)
	$(CONTROLLER_GEN) object:headerFile=$(BOILERPLATE_FILE) paths="./..."

	@$(OK) Generating Kubebuilder manifests

.PHONY: .do.kubebuilder.manifests
.kubebuilder.manifests.run: .do.kubebuilder.manifests

# ====================================================================================
# Common Targets

build.tools: $(KUBEBUILDER)
.test.init: $(KUBEBUILDER)
go.test.unit: $(KUBEBUILDER)

# ====================================================================================
# Special Targets

define KUBEBULDERV2_HELPTEXT
Kubebuilder Targets:
    kubebuilder.manifests   Generates Kubernetes custom resources manifests (e.g. CRDs RBACs, ...)

endef
export KUBEBULDERV2_HELPTEXT

.kubebuilder.help:
	@echo "$$KUBEBULDERV2_HELPTEXT"

.help: .kubebuilder.help
go.generate: kubebuilder.manifests

.PHONY: .kubebuilder.help kubebuilder.manifests

endif # __KUBEBUILDERV2_MAKEFILE__
