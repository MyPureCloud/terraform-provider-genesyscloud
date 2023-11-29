default: build

.PHONY: testacc clean build sideload

DIST_DIR=./dist
BIN_NAME=terraform-provider-genesyscloud
BIN_PATH=${DIST_DIR}/${BIN_NAME}

PLUGINS_DIR=~/.terraform.d/plugins
PLUGIN_PATH=genesys.com/mypurecloud/genesyscloud
DEV_VERSION=0.1.0

setup: copy-hooks

copy-hooks:
	chmod +x scripts/hooks/
	cp -r scripts/hooks .git/.

# Run acceptance tests
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -run TestAcc -timeout 120m -parallel 20  -coverprofile=coverage.out

# Run unit tests
testunit:
	TF_UNIT=1 go test ./... -run TestUnit -cover -count=1 -coverprofile=coverage_unit.out


coverageacc:
	go tool cover -func coverage.out | grep "total:" | \
	awk '{print ((int($$3) > 80) != 1) }'

coverageunit:
	go tool cover -func coverage_unit.out | grep "total:" | \
	awk '{print ((int($$3) > 80) != 1) }'

reportacc:
	go tool cover -html=coverage.out -o cover.html

reportunit:
	go tool cover -html=coverage_unit.out -o cover_unit.html

clean:
	rm -f -r ${DIST_DIR}
	rm -f -r ${PLUGINS_DIR}/${PLUGIN_PATH}
	rm -f -r ./.terraform

build:
	mkdir -p ${DIST_DIR}
	go build -o ${DIST_DIR} ./...

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)

sideload: build
	mkdir -p ${PLUGINS_DIR}/${PLUGIN_PATH}/${DEV_VERSION}/$(GOOS)_$(GOARCH)
	cp ${BIN_PATH} ${PLUGINS_DIR}/${PLUGIN_PATH}/${DEV_VERSION}/$(GOOS)_$(GOARCH)/${BIN_NAME}
