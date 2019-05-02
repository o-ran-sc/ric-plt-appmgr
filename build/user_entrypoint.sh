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

source /etc/os-release

if [ -z ${CONT_USER} ]; then CONT_USER=builder ; fi
if [ -z ${CONT_UID} ]; then CONT_UID=$(stat -c "%u" $(readlink -f .) ); fi
if [ -z ${CONT_GROUP} ]; then CONT_GROUP=builder; fi
if [ -z ${CONT_GID} ]; then CONT_GID=$(stat -c "%g" $(readlink -f .) ); fi

if [ $(id -u) -eq $CONT_UID ] || [ $(id -u) -ne 0 ] ; then
  exec "$@"
fi

if [ $(getent group ${CONT_GROUP}) ]  || [ $(getent group ${CONT_GID}) ] ; then
  echo "group conflict"
  exit 0 
fi
if [ $(getent passwd ${CONT_USER}) ] || [ $(getent passwd ${CONT_UID}) ] ; then
  echo "passwd conflict"
  exit 0 
fi

if [[ $ID == "ubuntu" ]] ; then
  groupadd --gid ${CONT_GID} ${CONT_GROUP} || exit 10
  useradd --shell /bin/bash --uid ${CONT_UID} --gid ${CONT_GID} -o -d /ws $(test -d /ws && echo "-M" || echo "-m") --groups $CONT_GID ${CONT_USER}|| exit 11
fi
if [[ $ID == "alpine" ]] ; then
  addgroup -g ${CONT_GID} ${CONT_GROUP}  || exit 10
  adduser -s /bin/bash -u ${CONT_UID} -G ${CONT_GROUP} -h /ws $(test -d /ws && echo "-H") -D ${CONT_USER} || exit 11
fi

chown ${CONT_UID}.${CONT_GID} /ws

echo "${CONT_USER} ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

DOCKER_SOCKET=/var/run/docker.sock
if [ -S ${DOCKER_SOCKET} ]; then
    DOCKER_GID=$(stat -c '%g' ${DOCKER_SOCKET})
    if [ $(getent group ${DOCKER_GID}) ]; then
      if [[ $ID == "ubuntu" ]] ; then
        usermod -aG $(getent group ${DOCKER_GID} | cut -d: -f1) ${CONT_USER}  || exit 12
      fi
      if [[ $ID == "alpine" ]] ; then
        addgroup ${CONT_USER} $(getent group ${DOCKER_GID} | cut -d: -f1)  || exit 12
      fi
    else
      if [[ $ID == "ubuntu" ]] ; then
        groupadd -for -g ${DOCKER_GID} docker_${CONT_USER}  || exit 13
        usermod -aG docker_${CONT_USER} ${CONT_USER}  || exit 14
      fi
      if [[ $ID == "alpine" ]] ; then
        addgroup -g ${DOCKER_GID} docker_${CONT_USER}  || exit 13
        addgroup ${CONT_USER} docker_${CONT_USER}  || exit 14
      fi
    fi
fi

export USER=${CONT_USER}
export HOME=/ws

mkdir -p /ws/go
chown -R ${CONT_UID}.${CONT_GID} /ws/go
export GOPATH="/ws/go"

sudo -E -s -u ${CONT_USER} env "PATH=$PATH" "$@"
