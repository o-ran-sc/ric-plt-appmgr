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

FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:2-u18.04-nng AS appmgr-build

RUN apt-get update -y && apt-get install -y jq

ENV PATH="/usr/local/go/bin:${PATH}"
ARG HELMVERSION=v2.12.3

# Install helm
RUN wget -nv https://storage.googleapis.com/kubernetes-helm/helm-${HELMVERSION}-linux-amd64.tar.gz \
    && tar -zxvf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && cp linux-amd64/helm /usr/local/bin/helm \
    && rm -rf helm-${HELMVERSION}-linux-amd64.tar.gz \
    && rm -rf linux-amd64

# Install kubectl from Docker Hub
COPY --from=lachlanevenson/k8s-kubectl:v1.10.3 /usr/local/bin/kubectl /usr/local/bin/kubectl

RUN mkdir -p /ws
WORKDIR "/ws"
ENV GOPATH="/go"

# Module prepare (if go.mod/go.sum updated)
COPY go.mod /ws
COPY go.sum /ws
RUN GO111MODULE=on go mod download

# build and test
COPY . /ws

RUN make -C /ws go-build

RUN make -C /ws go-test-fmt

#RUN make -C /ws go-test

CMD ["/bin/bash"]

#----------------------------------------------------------
FROM ubuntu:18.04 as appmgr

RUN apt-get update -y \
    && apt-get install -y sudo openssl ca-certificates ca-cacert \
    && apt-get clean

#
# libraries and helm
#
COPY --from=appmgr-build /usr/local/include/ /usr/local/include/
COPY --from=appmgr-build /usr/local/lib/ /usr/local/lib/
COPY --from=appmgr-build /usr/local/bin/helm /usr/local/bin/helm
COPY --from=appmgr-build /usr/local/bin/kubectl /usr/local/bin/kubectl

RUN ldconfig

#
# xApp Manager
#
RUN mkdir -p /opt/xAppManager \
    && chmod -R 755 /opt/xAppManager

COPY --from=appmgr-build /ws/cache/go/cmd/appmgr /opt/xAppManager/appmgr

WORKDIR /opt/xAppManager

COPY appmgr-entrypoint.sh /opt/xAppManager/
ENTRYPOINT ["/opt/xAppManager/appmgr-entrypoint.sh"]
