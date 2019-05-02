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
ifndef MAKE_DOCKER_TARGETS
MAKE_DOCKER_TARGETS:=1

.PHONY: docker-build docker-clean


docker-name_%:
	@echo $($*_DCKR_FULLNAME)

docker-build_%:
	@(\
		test -z "$${SSH_PRIVATE_KEY}" && SSH_PRIVATE_KEY=$$(cat $${HOME}/.ssh/id_rsa);\
		docker build --target $* $($*_DCKR_B_OPTS) --build-arg SSH_PRIVATE_KEY="$${SSH_PRIVATE_KEY}" -t $($*_DCKR_FULLNAME) -f $($*_DCKR_FILE) . ;\
	)

docker-irun_%:
	docker run $($*_DCKR_R_OPTS) $($*_DCKR_FULLNAME) /bin/bash

docker-irun-mounted_%:
	docker run $($*_DCKR_R_OPTS) -v $(shell pwd):/ws/go/src/${PACKAGEURL} --workdir "/ws/go/src/${PACKAGEURL}" $($*_DCKR_FULLNAME) /bin/bash

docker-run_%:
	docker run $($*_DCKR_R_OPTS) $($*_DCKR_FULLNAME)

docker-run-redished_%:
	@( \
		RETVAL=0;\
		docker network create --driver bridge $($*_DCKR_FULLNAME)-run_network;\
		docker run $($*_DCKR_R_OPTS) -d --name $($*_DCKR_FULLNAME)-run_redis --network $($*_DCKR_FULLNAME)-run_network redis;\
		docker run $($*_DCKR_R_OPTS) --name $($*_DCKR_FULLNAME)-run_xapp --network $($*_DCKR_FULLNAME)-run_network -e DBAAS_SERVICE_HOST=$($*_DCKR_FULLNAME)-run_redis $($*_DCKR_FULLNAME);\
		RETVAL=$$?;\
		docker stop $($*_DCKR_FULLNAME)-run_redis;\
		docker network rm $($*_DCKR_FULLNAME)-run_network;\
		exit $${RETVAL};\
	)
	
docker-clean_%:
	docker rmi $($*_DCKR_FULLNAME) || true


.SECONDEXPANSION:
docker-build: DCKR_TARGETS:=
docker-build: $$(DCKR_TARGETS)

.SECONDEXPANSION:
docker-clean: DCKR_TARGETS:=
docker-clean: $$(DCKR_TARGETS)

endif

#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

ifndef DCKR_FILE
DCKR_FILE:="Dockerfile"
endif

ifndef BUILD_PREFIX
BUILD_PREFIX:="${USER}-"
endif


#------------------------------------------------------------------------------
#
#------------------------------------------------------------------------------

ifndef $(DCKR_NAME)_DCKR_B_PREFIX
$(DCKR_NAME)_DCKR_B_PREFIX:=$(BUILD_PREFIX)
endif

ifndef $(DCKR_NAME)_DCKR_FILE
$(DCKR_NAME)_DCKR_FILE:=$(DCKR_FILE)
endif

$(DCKR_NAME)_DCKR_B_PREFIX:=$(subst /,_,$(shell echo $($(DCKR_NAME)_DCKR_B_PREFIX) | tr '[:upper:]' '[:lower:]'))

$(DCKR_NAME)_DCKR_FULLNAME:=$($(DCKR_NAME)_DCKR_B_PREFIX)$(DCKR_NAME)

$(DCKR_NAME)_DCKR_B_OPTS:=${DCKR_B_OPTS}
$(DCKR_NAME)_DCKR_B_OPTS:=$($(DCKR_NAME)_DCKR_B_OPTS) --network=host

ifndef PACKAGEURL
$(DCKR_NAME)_DCKR_B_OPTS:=$($(DCKR_NAME)_DCKR_B_OPTS) --build-arg PACKAGEURL=${PACKAGEURL}
endif

ifndef PACKAGEREPO
$(DCKR_NAME)_DCKR_B_OPTS:=$($(DCKR_NAME)_DCKR_B_OPTS) --build-arg PACKAGEREPO=${PACKAGEREPO}
endif

ifndef BUILD_PREFIX
$(DCKR_NAME)_DCKR_B_OPTS:=$($(DCKR_NAME)_DCKR_B_OPTS) --build-arg BUILD_PREFIX=${BUILD_PREFIX}
endif


$(DCKR_NAME)_DCKR_R_OPTS:=${DCKR_R_OPTS}
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS) --rm -i --net=host
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -t 0 && echo ' -t')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e /etc/localtime && echo ' -v /etc/localtime:/etc/localtime:ro')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e /var/run/docker.sock && echo ' -v /var/run/docker.sock:/var/run/docker.sock')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e ${HOME}/.docker && echo ' -v ${HOME}/.docker:/ws/.docker:ro')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e ${HOME}/.netrc && echo ' -v ${HOME}/.netrc:/ws/.netrc:ro')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e ${HOME}/.ssh && echo ' -v ${HOME}/.ssh:/ws/.ssh:ro')
$(DCKR_NAME)_DCKR_R_OPTS:=$($(DCKR_NAME)_DCKR_R_OPTS)$(shell test -e ${HOME}/.gitconfig && echo ' -v ${HOME}/.gitconfig:/ws/.gitconfig:ro')


docker-build: DCKR_TARGETS+=docker-build_$(DCKR_NAME)

docker-clean: DCKR_TARGETS+=docker-clean_$(DCKR_NAME)

