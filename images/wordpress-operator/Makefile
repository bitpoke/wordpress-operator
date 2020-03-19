PLATFORMS := linux_amd64
include ../../build/makelib/common.mk

IMAGE = $(BUILD_REGISTRY)/wordpress-operator-$(ARCH)
CACHE_IMAGES = $(IMAGE)
include ../../build/makelib/image.mk

img.build:
	@$(INFO) docker build $(IMAGE) $(PLATFORM)
	@cp -La . $(IMAGE_TEMP_DIR)
	@mkdir -p $(IMAGE_TEMP_DIR)/rootfs
	@cp $(OUTPUT_DIR)/bin/linux_$(ARCH)/wordpress-operator $(IMAGE_TEMP_DIR)/rootfs/dashboard
	@docker build $(BUILD_ARGS) \
		--build-arg ARCH=$(ARCH) \
		--build-arg TINI_VERSION=$(TINI_VERSION) \
		-t $(IMAGE) \
		$(IMAGE_TEMP_DIR)
	@$(OK) docker build $(IMAGE)
