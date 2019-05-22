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
FROM ubuntu:16.04 as appmgr-xapp-base

RUN apt-get update -y && \
    apt-get install -y wget

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
    autoconf-archive \
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
RUN mkdir -p /opt/build \
    && cd /opt/build && git clone https://gerrit.oran-osc.org/r/ric-plt/lib/rmr \
    && cd rmr/; mkdir build; cd build; cmake ..; make install \
    && cd /opt/build && git clone https://gerrit.oran-osc.org/r/com/log \
    && cd log/ ; ./autogen.sh ; ./configure ; make ; make install \
    && ldconfig

COPY build/user_entrypoint.sh /

RUN chmod +x /user_entrypoint.sh

RUN mkdir -p /ws
WORKDIR "/ws"
ENTRYPOINT ["/user_entrypoint.sh"]
CMD ["/bin/bash"]

#----------------------------------------------------------
#
#----------------------------------------------------------
FROM appmgr-xapp-base as appmgr-build

ARG PACKAGEURL
ARG HELMVERSION


#
# helm
#
RUN wget https://storage.googleapis.com/kubernetes-helm/helm-${HELMVERSION}-linux-amd64.tar.gz \
    && tar -zxvf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && cp linux-amd64/helm /usr/bin/helm \
    && rm -rf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && rm -rf linux-amd64

# Install kubectl from Docker Hub.
COPY --from=lachlanevenson/k8s-kubectl:v1.10.3 /usr/local/bin/kubectl /usr/local/bin/kubectl

RUN mkdir -p /go/src/${PACKAGEURL}
WORKDIR "/go/src/${PACKAGEURL}"
ENV GOPATH="/go"

# Module prepare (if go.mod/go.sum updated)
COPY go.mod /go/src/${PACKAGEURL}
COPY go.sum /go/src/${PACKAGEURL}
RUN GO111MODULE=on go mod download

# build
COPY . /go/src/${PACKAGEURL}
RUN make -C /go/src/${PACKAGEURL} build

CMD ["/bin/bash"]


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM appmgr-build as appmgr-test_unit
ARG PACKAGEURL
WORKDIR "/go/src/${PACKAGEURL}"
CMD ["make","go-test"]


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM appmgr-build as appmgr-test_fmt
ARG PACKAGEURL
WORKDIR "/go/src/${PACKAGEURL}"
CMD ["make","go-test-fmt"]

#----------------------------------------------------------
#
#----------------------------------------------------------
FROM appmgr-build as appmgr-test_sanity
ARG PACKAGEURL
WORKDIR "/go/src/${PACKAGEURL}"
CMD ["jq","-s",".", "api/appmgr_rest_api.json"]


#----------------------------------------------------------
#
#----------------------------------------------------------
FROM ubuntu:16.04 as appmgr
ARG PACKAGEURL

RUN apt-get update -y \
    && apt-get install -y sudo openssl ca-certificates ca-cacert \
    && apt-get clean


#
# libraries and helm
#
COPY --from=appmgr-build /usr/local/include/ /usr/local/include/
COPY --from=appmgr-build /usr/local/lib/ /usr/local/lib/
COPY --from=appmgr-build /usr/bin/helm /usr/bin/helm
COPY --from=appmgr-build /usr/local/bin/kubectl /usr/bin/kubectl

RUN ldconfig

#
# xApp
#
RUN mkdir -p /opt/xAppManager \
    && chmod -R 755 /opt/xAppManager

COPY --from=appmgr-build /go/src/${PACKAGEURL}/cache/go/cmd/appmgr /opt/xAppManager/appmgr
#COPY --from=appmgr-build /go/src/${PACKAGEURL}/config/appmgr.yaml /opt/etc/xAppManager/config-file.yaml


WORKDIR /opt/xAppManager

COPY appmgr-entrypoint.sh /opt/xAppManager/
ENTRYPOINT ["/opt/xAppManager/appmgr-entrypoint.sh"]
