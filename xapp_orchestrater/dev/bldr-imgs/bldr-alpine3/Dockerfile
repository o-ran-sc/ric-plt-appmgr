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

# This is an all-things-to-all-ORAN-people Alpine 3.11 build image with:
# gcc/g++ v 8
# golang v 1.13.x
# python 3.7
# ORAN libraries are not provided here to reduce version churn

FROM python:3.7-alpine
RUN apk update && apk add \
  autoconf \
  automake \
  build-base \
  cmake \
  libtool \
  ninja \
  pkgconfig \
  python-dev \
  sudo \
  git \
  go

# the autoconf-archive in alpine's edge repos is too new
# for our purposes, so we'll use one from the source.
RUN cd /tmp \
    && git clone --branch v2015.09.25 git://git.sv.gnu.org/autoconf-archive.git \
    && cp /tmp/autoconf-archive/m4/* /usr/share/aclocal \
    && rm -rf /tmp/autoconf-archive
