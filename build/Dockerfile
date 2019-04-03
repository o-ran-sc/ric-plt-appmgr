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

#----------------------------------------------------------
#
#----------------------------------------------------------
FROM ubuntu:16.04 as ubuntubase

RUN apt-get update -y && \
    apt-get install -y wget

#RUN sed -i -e 's,http://archive.ubuntu.com/ubuntu,mirror://mirrors.ubuntu.com/mirrors.txt,' /etc/apt/sources.list
#RUN sed -i -e 's,http://security.ubuntu.com/ubuntu,mirror://mirrors.ubuntu.com/mirrors.txt,' /etc/apt/sources.list
#RUN sed -i -e 's,http://archive.ubuntu.com/ubuntu,http://mirrors.nic.funet.fi/ubuntu,' /etc/apt/sources.list
#RUN sed -i -e 's,http://security.ubuntu.com/ubuntu,http://mirrors.nic.funet.fi/ubuntu,' /etc/apt/sources.list

RUN sed -i -e "s,http://archive.ubuntu.com/ubuntu,$(wget -qO - mirrors.ubuntu.com/mirrors.txt | head -1)," /etc/apt/sources.list
RUN sed -i -e "s,http://security.ubuntu.com/ubuntu,$(wget -qO - mirrors.ubuntu.com/mirrors.txt | head -1)," /etc/apt/sources.list

#
# packages
#
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y \
    build-essential \
    apt-utils \
    cmake \
    make \
    autoconf \
    gawk \
    libtool \
    automake \
    pkg-config \
    sudo \
    wget \
    nano \
    git \
    jq
 

#
# go
#
RUN wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz && \
	tar -C /usr/local -xvf ./go1.12.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

#
# rancodev libs
#
RUN echo "137.135.91.204 gerrit.ranco-dev-tools.eastus.cloudapp.azure.com" >> /etc/hosts \
    && mkdir -p /opt/build \
    && cd /opt/build && git clone http://gerrit.ranco-dev-tools.eastus.cloudapp.azure.com/ric-plt/lib/rmr \
    && cd rmr/; mkdir build; cd build; cmake ..; make install \
    && cd /opt/build && git clone -b v0.0.1 http://gerrit.ranco-dev-tools.eastus.cloudapp.azure.com/com/log \
    && cd log/ ; ./autogen.sh ; ./configure ; make ; make install \
    && ldconfig

#    && cd /opt/build && git clone http://gerrit.ranco-dev-tools.eastus.cloudapp.azure.com/ric-plt/lib/rmr \
#    && cd rmr/; mkdir build; cd build; cmake ..; make install \


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM ubuntubase as builder

ARG HELMVERSION

#
# helm
#
RUN wget https://storage.googleapis.com/kubernetes-helm/helm-${HELMVERSION}-linux-amd64.tar.gz \
    && tar -zxvf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && cp linux-amd64/helm /usr/bin/helm \
    && rm -rf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && rm -rf linux-amd64


#
# xapp_manager codes
#
RUN mkdir -p /go/src/appmgr
ENV GOPATH="/go"

#
# Speed up things by generating layer with needed go packages
#
RUN go get github.com/gorilla/mux \
    && go get github.com/spf13/viper \
    && go get github.com/gorilla/mux \
    && go get github.com/orcaman/concurrent-map \
    && go get github.com/segmentio/ksuid \
    && go get gopkg.in/yaml.v2


COPY . /go/src/appmgr


#
# build
#
RUN make -C /go/src/appmgr/build deps

RUN make -C /go/src/appmgr/build build


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM builder as test_unit
WORKDIR "/go/src/appmgr"
CMD ["make","-C","build", "unit-test"]


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM builder as test_sanity
WORKDIR "/go/src/appmgr"
CMD ["jq","-s",".", "rest_api/xapp_manager_rest_api.json"]

#----------------------------------------------------------
#
#----------------------------------------------------------
FROM ubuntu:16.04 as release

RUN apt-get update -y \
    && apt-get install -y sudo openssl ca-certificates ca-cacert \
    && apt-get clean


#
# libraries and helm
#
COPY --from=builder /usr/local/include/ /usr/local/include/
COPY --from=builder /usr/local/lib/ /usr/local/lib/
COPY --from=builder /usr/bin/helm /usr/bin/helm

RUN ldconfig

#
# xApp
#
RUN mkdir -p /opt/xAppManager \
    && chmod -R 755 /opt/xAppManager

COPY --from=builder /go/src/appmgr/build/appmgr /opt/xAppManager/appmgr
#COPY --from=builder /go/src/appmgr/config/xapp-manager.yaml /opt/etc/xAppManager/config-file.yaml


COPY build/docker-entrypoint.sh /opt/xAppManager/

WORKDIR /opt/xAppManager

ENTRYPOINT ["/opt/xAppManager/docker-entrypoint.sh"]
