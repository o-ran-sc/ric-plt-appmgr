
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

.DEFAULT: go-build

default: go-build

build: go-build

test: go-test

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------
ROOT_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
CACHE_DIR?=$(abspath $(ROOT_DIR)/cache)

#------------------------------------------------------------------------------
#
# Build and test targets
#
#------------------------------------------------------------------------------

XAPP_NAME:=appmgr
XAPP_ROOT:=cmd
XAPP_TESTENV:="RMR_SEED_RT=config/uta_rtg.rt CFG_FILE=$(ROOT_DIR)helm_chart/uemgr/descriptors/config-file.json"
include build/make.go.mk 

#------------------------------------------------------------------------------
#
# DOCKER TARGETS
#
#------------------------------------------------------------------------------

HELMVERSION:=v2.13.0-rc.1
DCKR_B_OPTS:=${DCKR_B_OPTS} --build-arg HELMVERSION=${HELMVERSION} 

PACKAGEURL:="gerrit.oran-osc.org/r/ric-plt/appmgr"

DCKR_NAME:=appmgr-xapp-base
include build/make.docker.mk

DCKR_NAME:=appmgr-test_unit
include build/make.docker.mk

DCKR_NAME:=appmgr-test_fmt
include build/make.docker.mk

DCKR_NAME:=appmgr-test_sanity
include build/make.docker.mk

DCKR_NAME:=appmgr
include build/make.docker.mk


docker-test: docker-run-stop_appmgr-test_fmt docker-run-stop_appmgr-test_sanity docker-run-stop_appmgr-test_unit

