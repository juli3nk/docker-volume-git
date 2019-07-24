REPO_NAME = kassisol/docker-volume-git
PLUGIN_NAME = kassisol/gitvol
PLUGIN_TAG ?= latest
DEV_IMAGE_NAME = juliengk/dev:go

.PHONY: all
all: clean rootfs create

.PHONY: clean
clean:
	@echo "### rm ./build"
	@rm -rf ./build

.PHONY: dev
dev:
	docker container run -ti --rm --mount type=bind,src=$$PWD,dst=/go/src/github.com/${REPO_NAME} --workdir /go/src/github.com/${REPO_NAME} --name docker-volume-git-dev ${DEV_IMAGE_NAME}

.PHONY: config
config:
	@echo "### copy config.json to ./build/"
	@mkdir -p ./build
	@cp config.json ./build/

.PHONY: rootfs
rootfs: config
	@echo "### docker build: rootfs image with"
	@docker image build -t ${PLUGIN_NAME}:rootfs .
	@echo "### create rootfs directory in ./build/rootfs"
	@mkdir -p ./build/rootfs
	@docker create --name tmp ${PLUGIN_NAME}:rootfs
	@docker export tmp | tar -x -C ./build/rootfs
	@docker rm -vf tmp

.PHONY: create
create:
	@echo "### remove existing plugin ${PLUGIN_NAME}:${PLUGIN_TAG} if exists"
	@docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo "### create new plugin ${PLUGIN_NAME}:${PLUGIN_TAG} from ./build"
	@docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./build

.PHONY: enable
enable:
	@echo "### enable plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	@docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

.PHONY: disable
disable:
	@echo "### disable plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	@docker plugin disable ${PLUGIN_NAME}:${PLUGIN_TAG}

.PHONY: push
push:  clean rootfs create enable
	@echo "### push plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	@docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
