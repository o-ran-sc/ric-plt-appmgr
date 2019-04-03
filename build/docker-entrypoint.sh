#!/bin/bash

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

cp /opt/ric/config/xapp-manager.yaml /opt/xAppManager/config-file.yaml

# Update helm repo username and password to the configuration yaml, so all configuration can be parsed centrally
HELM_USERNAME=$(cat /opt/ric/secret/helm_repo_username) && HELM_PASSWORD=$(cat /opt/ric/secret/helm_repo_password)
echo '"helm_username": '$HELM_USERNAME >> /opt/xAppManager/config-file.yaml
echo '"helm_password": '$HELM_PASSWORD >> /opt/xAppManager/config-file.yaml

# Copy all certificates from mounted folder to root system
cp /opt/ric/certificates/* /etc/ssl/certs

# Start services, etc.
/opt/xAppManager/appmgr -f /opt/xAppManager/config-file.yaml
