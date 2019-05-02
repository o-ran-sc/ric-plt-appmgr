#!/bin/bash
# Copyright (c) 2019 AT&T Intellectual Property.
# Copyright (c) 2019 Nokia.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# SSH
#
if [ -n  "${SSH_PRIVATE_KEY}" ] && [ -n  "${PACKAGEREPO}" ] ; then

    # ssh configs (no private key)
	mkdir -p ${HOME}/.ssh
	test -n "${PACKAGEREPO}" && ssh-keyscan -H ${PACKAGEREPO} >> ${HOME}/.ssh/known_hosts
	echo -e "IdentityFile ~/.ssh/id_rsa\n" > ${HOME}/.ssh/config
	echo -e "Host *\n\tStrictHostKeyChecking no\n\n" >> ${HOME}/.ssh/config
	chmod -R go= ${HOME}/.ssh/
	chmod -R u+rw ${HOME}/.ssh/

	# gitconfig
	test -n "${PACKAGEREPO}" && echo -e "[url \"ssh://git@${PACKAGEREPO}/\"]\n\tinsteadOf = https://${PACKAGEREPO}/" > ${HOME}/.gitconfig

	# ssh agent
    TEMPFILE=/dev/shm/deployment.key
    echo "$SSH_PRIVATE_KEY" > $TEMPFILE
    chmod 0600 $TEMPFILE
    eval $(ssh-agent)
    ssh-add $TEMPFILE
    rm $TEMPFILE
fi

if [ -n  "${NETRC_CONFIG}" ] ; then
	echo "${NETRC_CONFIG}"  > /root/.netrc
fi


exec $*
