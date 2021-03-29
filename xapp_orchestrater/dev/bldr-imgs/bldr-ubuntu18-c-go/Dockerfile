# O-RAN-SC
#
# Copyright (C) 2019-2020 AT&T Intellectual Property and Nokia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Builder image with superset of CI tools for RIC components

FROM nexus3.o-ran-sc.org:10001/ubuntu:18.04

RUN apt-get update && apt-get install -y \
    alien \
    autoconf \
    autoconf-archive \
    automake \
    autotools-dev \
    build-essential \
    cmake \
    g++  \
    gcc \
    git \
    ksh \
    libbz2-dev \
    libicu-dev \
    libsctp-dev \
    libtool \
    lksctp-tools \
    make \
    python-dev \
    pkg-config \
    software-properties-common \
    sudo \
    wget \
    zlib1g \
    zlib1g-dev \
    zlibc \
    zip

WORKDIR /tmp

# Install go versions in /opt
RUN mkdir -p /opt/go

ARG g12="1.12.17"
RUN wget -nv https://dl.google.com/go/go${g12}.linux-amd64.tar.gz \
    && tar -xf go${g12}.linux-amd64.tar.gz \
    && mv go /opt/go/1.12 \
    && rm -f go*.gz

ARG g13="1.13.12"
RUN wget -nv https://dl.google.com/go/go${g13}.linux-amd64.tar.gz \
    && tar -xf go${g13}.linux-amd64.tar.gz \
    && mv go /opt/go/1.13 \
    && rm -f go*.gz

ARG g14="1.14.4"
RUN wget -nv https://dl.google.com/go/go${g14}.linux-amd64.tar.gz \
    && tar -xf go${g14}.linux-amd64.tar.gz \
    && mv go /opt/go/1.14 \
    && rm -f go*.gz

ENV GOPATH=/root/go
# Default path has go version 1.12.x
# Use /opt/go/1.13/bin for version 1.13.x
# Use /opt/go/1.14/bin for version 1.14.x
ENV PATH=$PATH:/opt/go/1.12/bin:$GOPATH/bin

# This installs go-acc to GOPATH
RUN go version \
    && go get -v github.com/ory/go-acc \
    && go-acc --help

# install a recent version of cmake, ubuntu repo version is old
ARG cv="3.14.5"
RUN wget -nv https://github.com/Kitware/CMake/releases/download/v${cv}/cmake-${cv}-Linux-x86_64.sh \
    && chmod +x cmake-${cv}-Linux-x86_64.sh \
    && ./cmake-${cv}-Linux-x86_64.sh --prefix=/usr/local --skip-license \
    && rm -f cmake*.sh

# sonar-scanner is not available in ubuntu repo
ARG sv="4.2.0.1873"
RUN wget -nv https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-${sv}-linux.zip \
    && unzip sonar-scanner-cli-${sv}-linux.zip \
    && mv sonar-scanner-${sv}-linux /usr/local/sonar-scanner \
    && rm -f sonar*.zip

# golangci-lint is not available in ubuntu repo
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOPATH/bin v1.23.6 \
    && rm -f install.sh

# install a recent version of boost, ubuntu repo version is old
ARG bv="1_69_0"
RUN wget -nv https://dl.bintray.com/boostorg/release/1.69.0/source/boost_${bv}.zip \
    && unzip -q boost_${bv}.zip \
    && cd boost_${bv} \
    && ./bootstrap.sh --prefix=/usr/ \
    && ./b2 \
    && ./b2 install \
    && cd ../ \
    && rm -rf boost*.zip boost_${bv}
