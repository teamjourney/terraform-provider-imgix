HOSTNAME = journey.travel
VERSION = 1.0.0
NAMESPACE = terraform-providers
NAME = imgix
BINARY = terraform-provider-${NAME}
OS_ARCH = linux_amd64

build:
	go build -o ${BINARY}
	mkdir -p ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

init:
	cd example && \
	rm -rf .terraform && \
	terraform0.13 init

apply: build init
	cd example && \
	terraform0.13 apply

import: build init
	cd example && \
	(terraform0.13 state rm imgix_source.this || true) && \
	terraform0.13 import imgix_source.this 60081817b721f6b6a15b09e7
