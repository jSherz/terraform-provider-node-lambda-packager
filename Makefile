TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=jsherz
NAME=node-lambda-packager
BINARY=terraform-provider-${NAME}
VERSION=1.5.3
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --clean --snapshot --skip publish,sign

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/${HOSTNAME}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/registry.terraform.io/${HOSTNAME}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
