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
#------------------------------------------------------------------------------
#ROOT_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

ifndef ROOT_DIR
$(error ROOT_DIR NOT DEFINED)
endif
BUILD_DIR?=$(abspath $(ROOT_DIR)/build)
CACHE_DIR?=$(abspath $(BUILD_DIR)/cache)


#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------
ifndef MAKE_GO_TARGETS
MAKE_GO_TARGETS:=1


GOOS=$(shell go env GOOS)
GOCMD=go
GOBUILD=$(GOCMD) build -a -installsuffix cgo
GORUN=$(GOCMD) run -a -installsuffix cgo
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v
GOGET=$(GOCMD) get

GOFILES:=$(shell find $(ROOT_DIR) -name '*.go' -not -name '*_test.go')
GOALLFILES:=$(shell find $(ROOT_DIR) -name '*.go')
GOMODFILES:=go.mod go.sum

.PHONY: FORCE go-build go-test go-test-fmt go-fmt go-clean
 

FORCE:


$(CACHE_DIR)/%: $(GOFILES) $(GOMODFILES)
	@echo "Building:\t$*"
	GO111MODULE=on GO_ENABLED=0 GOOS=linux $(GOBUILD) -o $@ ./$*


$(CACHE_DIR)/%_test: $(GOALLFILES) $(GOMODFILES)
	@echo "Testing:\t$*"
	GO111MODULE=on GO_ENABLED=0 GOOS=linux $(GOTEST) -coverprofile $(COVEROUT) -c -o $@ ./$*
	test -e $@ && (eval $(TESTENV) $@ -test.coverprofile $(COVEROUT) || false) || true
	test -e $@ && (go tool cover -html=$(COVEROUT) -o $(COVERHTML) || false) || true


.SECONDEXPANSION:
go-build: XAPP_TARGETS:=
go-build: $$(XAPP_TARGETS)

.SECONDEXPANSION:
go-test: XAPP_TARGETS:=
go-test: go-clean $$(XAPP_TARGETS)

go-test-fmt: $(GOFILES)
	@(RESULT="$$(gofmt -l $^)"; test -z "$${RESULT}" || (echo -e "gofmt failed:\n$${RESULT}" && false) )

go-fmt: $(GOFILES)
	gofmt -w -s $^

go-clean: XAPP_TARGETS:=
go-clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf $(XAPP_TARGETS)* 2> /dev/null
	go clean 2> /dev/null


endif

#------------------------------------------------------------------------------
#
#-------------------------------------------------------------------- ----------

$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_test: COVEROUT:=$(abspath $(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_cover.out)
$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_test: COVERHTML:=$(abspath $(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_cover.html)
$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_test: TESTENV:=$(XAPP_TESTENV)

go-build: XAPP_TARGETS+=$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)
go-test: XAPP_TARGETS+=$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_test
go-clean: XAPP_TARGETS+=$(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME) $(CACHE_DIR)/$(XAPP_ROOT)/$(XAPP_NAME)_test

