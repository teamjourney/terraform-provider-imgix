HOSTNAME = journey.travel
VERSION = 1.0.0
NAMESPACE = terraform-providers
NAME = imgix
BINARY = terraform-provider-${NAME}
OS_ARCH = linux_amd64

build:
	go build -o ${BINARY}
	mkdir -p ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/
