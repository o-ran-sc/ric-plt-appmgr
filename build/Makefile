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

BUILD_DIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

ROOT_DIR:=$(abspath $(BUILD_DIR)/..)

BUILD_PREFIX?="${USER}-"

XAPP_MGR:=appmgr
XAPP_MGR_DOCKER:=${BUILD_PREFIX}xapp-manager

GOSRC := $(abspath $(BUILD_DIR)/../src)
GOFILES := $(GOSRC)/*.go
COVEROUT := $(abspath $(BUILD_DIR)/cover.out)
COVERHTML := $(abspath $(BUILD_DIR)/cover.html)

GOCMD=go
GOBUILD=$(GOCMD) build -a -installsuffix cgo
GORUN=$(GOCMD) run -a -installsuffix cgo
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v -coverprofile $(COVEROUT)
GOGET=$(GOCMD) get

HELMVERSION:=v2.13.0-rc.1

#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------
.PHONY: FORCE build deps run unit-test test-pkg test clean docker-base-build docker-base-clean docker-build docker-run docker-clean docker-test-build docker-test-run-unittest docker-test-run-sanity docker-test-run docker-test-clean
 
.DEFAULT: build

default: build

FORCE:

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

XAPP_MGR_DOCKER:=$(shell echo $(XAPP_MGR_DOCKER) | tr '[:upper:]' '[:lower:]')

#XAPP_MGR_DOCKER:=$(subst /,_,${XAPP_MGR_DOCKER})

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

$(BUILD_DIR)$(XAPP_MGR): deps ${wildcard $(GOFILES)}
	GO_ENABLED=0 GOOS=linux $(GOBUILD) -o $(BUILD_DIR)$(XAPP_MGR) $(GOFILES)

build: $(BUILD_DIR)$(XAPP_MGR)

deps: ${wildcard $(GOFILES)}
	cd $(GOSRC) && $(GOGET)

run: $(BUILD_DIR)$(XAPP_MGR)
	$(BUILD_DIR)$(XAPP_MGR) -host-addr="localhost:8080" -helm-host="localhost:31807" -helm-chart="./"

unit-test:
	cd $(GOSRC) && $(GOTEST)
	go tool cover -html=$(COVEROUT) -o $(COVERHTML)

clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf $(XAPP_MGR) 2> /dev/null
	go clean 2> /dev/null

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

DCKR_BUILD_OPTS:=--network=host --build-arg HELMVERSION=${HELMVERSION}

DCKR_RUN_OPTS:=--rm -i
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -t 0 && echo ' -t')
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -e /etc/localtime && echo ' -v /etc/localtime:/etc/localtime:ro')
DCKR_RUN_OPTS:=${DCKR_RUN_OPTS}$(shell test -e /var/run/docker.sock && echo ' -v /var/run/docker.sock:/var/run/docker.sock')


#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------
docker-name:
	@echo $(XAPP_MGR_DOCKER)

docker-build:
	docker build --target release ${DCKR_BUILD_OPTS} -t $(XAPP_MGR_DOCKER) -f Dockerfile ../.

docker-run:
	docker run ${DCKR_RUN_OPTS} -v /opt/ric:/opt/ric -p 8080:8080 $(XAPP_MGR_DOCKER)

docker-clean:
	docker rmi $(XAPP_MGR_DOCKER)


#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

docker-test-build:
	docker build --target test_unit ${DCKR_BUILD_OPTS} -t ${XAPP_MGR_DOCKER}-test_unit -f Dockerfile ../.
	docker build --target test_sanity ${DCKR_BUILD_OPTS} -t ${XAPP_MGR_DOCKER}-test_sanity -f Dockerfile ../.

docker-test-run-unit:
	docker run ${DCKR_RUN_OPTS} ${XAPP_MGR_DOCKER}-test_unit 

docker-test-run-sanity:
	docker run ${DCKR_RUN_OPTS} ${XAPP_MGR_DOCKER}-test_sanity

docker-test-run: docker-test-run-sanity docker-test-run-unit

docker-test-clean:
	docker rmi -f ${XAPP_MGR_DOCKER}-test_unit
	docker rmi -f ${XAPP_MGR_DOCKER}-test_sanity

