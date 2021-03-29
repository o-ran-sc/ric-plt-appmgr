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

# Provides MDCLOG for Alpine including header files and shared-object libraries,
# basically a workaround for missing APK support at PackageCloud.io.
# Unlike other bldr- images, this only has library files, no build tools.
# See list of available files below.  Use the following command to copy a file
# in a Docker build, after replacing "TAG" with the desired version string:
#
# COPY --from=nexus3.o-ran-sc.org:10002/bldr-alpine3-mdclog:TAG /usr/local/lib/libmdclog.so /usr/local/lib/libmdclog.so
#
# /usr/local/include/mdclog:
#     mdclog.h
#
# /usr/local/lib:
#     libmdclog.a
#     libmdclog.la
#     libmdclog.so
#     libmdclog.so.0         (subject to version change)
#     libmdclog.so.0.0.4     (ditto)

FROM alpine:3.11 as bldr
RUN apk update && apk add autoconf automake build-base cmake libtool pkgconfig git sudo

# the autoconf-archive in alpine's edge repos is too new
# for our purposes, so we'll use one from the source.
RUN cd /tmp \
    && git clone --branch v2015.09.25 git://git.sv.gnu.org/autoconf-archive.git \
    && cp /tmp/autoconf-archive/m4/* /usr/share/aclocal \
    && rm -rf /tmp/autoconf-archive

# MCDLOG, version is pinned
RUN cd /tmp \
   && git clone --branch v0.0.4 https://gerrit.o-ran-sc.org/r/com/log \
   && cd log \
   && ./autogen.sh \
   && ./configure \
   && make \
   && make install

FROM scratch
COPY --from=bldr /usr/local/include/mdclog /usr/local/include/mdclog
COPY --from=bldr /usr/local/lib/libmdclog.* /usr/local/lib/
