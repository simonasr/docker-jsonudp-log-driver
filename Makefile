# Plugin parameters
PLUGIN_NAME=simonasr/jsonudpdriver
PLUGIN_BUILD_DIR=/tmp/.plugin-build
PLUGIN_TAG=0.1.0

all: clean build-image create-plugin enable-plugin

clean:
	@echo "Removing plugin build directory"
	rm -rf ${PLUGIN_BUILD_DIR}

build-image:
	@echo "docker build: rootfs image with the plugin"
	docker build -f Dockerfile -t ${PLUGIN_NAME}:${PLUGIN_TAG} .
	@echo "### create rootfs directory in ${PLUGIN_BUILD_DIR}/rootfs"
	mkdir -p ${PLUGIN_BUILD_DIR}/rootfs
	docker create --name rootfsctr ${PLUGIN_NAME}:${PLUGIN_TAG}
	docker export rootfsctr | tar -x -C ${PLUGIN_BUILD_DIR}/rootfs
	@echo "### copy config.json to ${PLUGIN_BUILD_DIR}/"
	cp config.json ${PLUGIN_BUILD_DIR}/
	docker rm -vf rootfsctr

create-plugin:
	@echo "### remove existing plugin ${PLUGIN_NAME}:${PLUGIN_TAG} if exists"
	docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo "### create new plugin ${PLUGIN_NAME}:${PLUGIN_TAG} from ${PLUGIN_BUILD_DIR}"
	docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ${PLUGIN_BUILD_DIR}

enable-plugin:
	@echo "### enable plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

push: clean build-image create-plugin enable-plugin
	@echo "### push plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
