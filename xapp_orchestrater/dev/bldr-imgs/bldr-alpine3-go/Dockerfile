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

# Has:
# golang 1.13.8
# alpine 3.11
# rmr 3.x

# install a well known working rmr
FROM golang:1.13.8-alpine3.11
RUN apk update && apk add autoconf automake build-base cmake libtool ninja pkgconfig git sudo
RUN git clone --branch 3.7.1 https://gerrit.oran-osc.org/r/ric-plt/lib/rmr \
    && cd rmr \
    && mkdir .build; cd .build \
    && echo "<<<installing rmr devel headers>>>" \
    && cmake .. -DDEV_PKG=1; make install \
    && echo "<<< installing rmr .so>>>" \
    && cmake .. -DDEV_PKG=0; make install
