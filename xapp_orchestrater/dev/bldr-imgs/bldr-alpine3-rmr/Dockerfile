# ==================================================================================
#       Copyright (c) 2020 Nokia
#       Copyright (c) 2020 AT&T Intellectual Property.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#          http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
# ==================================================================================

# Provides RMR for Alpine including header files and shared-object libraries,
# basically a workaround for missing APK support at PackageCloud.io.
# Unlike other bldr- images, this only has library files, no build tools.
# See list of available files below.  Use the following command to copy a file
# in a Docker build, after replacing "TAG" with the desired version string:
#
# COPY --from=nexus3.o-ran-sc.org:10002/o-ran-sc/bldr-alpine3-rmr:TAG /usr/local/lib64/librmr* /usr/local/lib64/
#
# /usr/local/bin:
#     rmr_probe
#
# /usr/local/include/rmr:
#     RIC_message_types.h
#     rmr.h
#     rmr_logging.h
#     rmr_symtab.h
#
# /usr/local/lib64:
#     librmr_si.a
#     librmr_si.so
#     librmr_si.so.4         (subject to version change)
#     librmr_si.so.4.0.2     (subject to version change)

FROM alpine:3.11 as bldr
RUN apk update && apk add autoconf automake build-base cmake libtool pkgconfig git sudo
RUN git clone --branch 4.1.2 https://gerrit.oran-osc.org/r/ric-plt/lib/rmr \
    && cd rmr \
    && mkdir .build; cd .build \
    && echo "<<<installing rmr devel headers>>>" \
    && cmake .. -DDEV_PKG=1; make install \
    && echo "<<< installing rmr .so>>>" \
    && cmake .. -DDEV_PKG=0; make install

FROM scratch
COPY --from=bldr /usr/local/bin/rmr* /usr/local/bin/
COPY --from=bldr /usr/local/include/rmr /usr/local/include/rmr
COPY --from=bldr /usr/local/lib64/librmr* /usr/local/lib64/
