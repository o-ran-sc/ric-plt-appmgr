#   Copyright (c) 2019 AT&T Intellectual Property.
#   Copyright (c) 2019 Nokia.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------
ROOT_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR:=$(abspath $(ROOT_DIR)/build)

PACKAGEURL:="gerrit.oran-osc.org/r/ric-plt/appmgr"
HELMVERSION:=v2.13.0-rc.1

#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------
COVEROUT := $(abspath $(BUILD_DIR)/cover.out)
COVERHTML := $(abspath $(BUILD_DIR)/cover.html)

GOOS=$(shell go env GOOS)
GOCMD=go
GOBUILD=$(GOCMD) build -a -installsuffix cgo
GORUN=$(GOCMD) run -a -installsuffix cgo
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v -coverprofile $(COVEROUT)
GOGET=$(GOCMD) get

GOFILES := $(shell find $(ROOT_DIR) -name '*.go' -not -name '*_test.go')  go.mod go.sum
GOFILES_NO_VENDOR := $(shell find $(ROOT_DIR) -path ./vendor -prune -o -name "*.go" -not -name '*_test.go' -print)

CMDS:=$(BUILD_DIR)/appmgr

#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------
 .DEFAULT: build

default: build

.PHONY: FORCE 

FORCE:

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

$(CMDS): $(GOFILES)
	GO111MODULE=on GO_ENABLED=0 GOOS=linux $(GOBUILD) -o $@ ./cmd/$(shell basename "$@")


$(addsuffix _test,$(CMDS)): $(GOFILES)
	GO111MODULE=on GO_ENABLED=0 GOOS=linux $(GOTEST) -c -o $@ ./cmd/$(patsubst %_test,%, $(shell basename "$@")) 
	timeout -s KILL 5s $@ -test.coverprofile $(COVEROUT)
	go tool cover -html=$(COVEROUT) -o $(COVERHTML)


build: $(CMDS)


test: $(addsuffix _test,$(CMDS))


test-fmt: $(GOFILES_NO_VENDOR)
	@(RESULT="$$(gofmt -l $^)"; test -z "$${RESULT}" || (echo -e "gofmt failed:\n$${RESULT}" && false) )


fmt: $(GOFILES_NO_VENDOR)
	gofmt -w -s $^


clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf $(CMDS)* 2> /dev/null
	go clean 2> /dev/null

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

BUILD_PREFIX?="${USER}-"

DCKR_FILE:=docker/Dockerfile

DCKR_NAME:=${BUILD_PREFIX}appmgr
DCKR_NAME:=$(shell echo $(DCKR_NAME) | tr '[:upper:]' '[:lower:]')
DCKR_NAME:=$(subst /,_,${DCKR_NAME})

DCKR_BUILD_OPTS:=${DCKR_BUILD_OPTS} --network=host --build-arg HELMVERSION=${HELMVERSION} --build-arg PACKAGEURL=${PACKAGEURL}

DCKR_RUN_OPTS:=${DCKR_RUN_OPTS} --rm -i
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -t 0 && echo ' -t')
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -e /etc/localtime && echo ' -v /etc/localtime:/etc/localtime:ro')
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -e /var/run/docker.sock && echo ' -v /var/run/docker.sock:/var/run/docker.sock')


#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------
docker-name:
	@echo $(DCKR_NAME)

docker-build:
	docker build --target release ${DCKR_BUILD_OPTS} -t $(DCKR_NAME) -f $(DCKR_FILE) .

docker-run:
	docker run ${DCKR_RUN_OPTS} -v /opt/ric:/opt/ric -p 8080:8080 $(DCKR_NAME)

docker-clean:
	docker rmi $(DCKR_NAME)


#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

docker-test-build:
	docker build --target test_unit ${DCKR_BUILD_OPTS} -t ${DCKR_NAME}-test_unit -f $(DCKR_FILE) .
	docker build --target test_sanity ${DCKR_BUILD_OPTS} -t ${DCKR_NAME}-test_sanity -f $(DCKR_FILE) .
	docker build --target test_fmt ${DCKR_BUILD_OPTS} -t ${DCKR_NAME}-test_fmt -f $(DCKR_FILE) .

docker-test-run-unit:
	@( \
		RETVAL=0;\
		docker network create --driver bridge ${DCKR_NAME}-test_unit_network;\
		docker run ${DCKR_RUN_OPTS} -d --name ${DCKR_NAME}-test_unit_redis --network ${DCKR_NAME}-test_unit_network redis;\
		docker run ${DCKR_RUN_OPTS} --name ${DCKR_NAME}-test_unit_run --network ${DCKR_NAME}-test_unit_network -e DBAAS_SERVICE_HOST=${DCKR_NAME}-test_unit_redis ${DCKR_NAME}-test_unit;\
		RETVAL=$$?;\
		docker stop ${DCKR_NAME}-test_unit_redis;\
		docker network rm ${DCKR_NAME}-test_unit_network;\
		exit $${RETVAL};\
	)


docker-test-run-fmt:
	docker run ${DCKR_RUN_OPTS} ${DCKR_NAME}-test_fmt

docker-test-run-sanity:
	docker run ${DCKR_RUN_OPTS} ${DCKR_NAME}-test_sanity

docker-test-clean:
	docker rmi -f ${DCKR_NAME}-test_unit
	docker rmi -f ${DCKR_NAME}-test_sanity
	docker rmi -f ${DCKR_NAME}-test_fmt

