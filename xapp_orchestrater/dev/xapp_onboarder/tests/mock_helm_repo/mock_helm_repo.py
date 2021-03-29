#!/usr/bin/env python3
################################################################################
#   Copyright (c) 2020 AT&T Intellectual Property.                             #
#                                                                              #
#   Licensed under the Apache License, Version 2.0 (the "License");            #
#   you may not use this file except in compliance with the License.           #
#   You may obtain a copy of the License at                                    #
#                                                                              #
#       http://www.apache.org/licenses/LICENSE-2.0                             #
#                                                                              #
#   Unless required by applicable law or agreed to in writing, software        #
#   distributed under the License is distributed on an "AS IS" BASIS,          #
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   #
#   See the License for the specific language governing permissions and        #
#   limitations under the License.                                             #
################################################################################

import pkg_resources
import logging.config
import yaml
import json
import os
import tarfile
from flask import Flask, jsonify, send_file
from tests.constants import config_file, controls_schema_file, helm_repo_index_response
from xapp_onboarder.server import settings


logger_config = pkg_resources.resource_filename("xapp_onboarder", 'logging.conf')
logging.config.fileConfig(logger_config)
log = logging.getLogger(__name__)

listen_address = '0.0.0.0:8080'
app = Flask(__name__)
app.config['SERVER_NAME'] = listen_address




@app.route('/')
def root_dir():
    return {'health': 'OK'}

@app.route('/index.yaml')
def get_index():
    return yaml.dump(helm_repo_index_response)

@app.route('/api/charts', methods=['POST'])
def upload_chart():
    return {"saved": True}, 201

@app.route('/api/charts/test_xapp/1.0.0', methods=['DELETE'])
def delete_chart():
    return {"deleted":True}, 200

@app.route('/api/charts/test_xapp', methods=['GET'])
def get_all_xapp_version():
    test_xapp = helm_repo_index_response['entries']['test_xapp']
    return jsonify(test_xapp), 200

@app.route('/api/charts', methods=['GET'])
def get_all_xapp():
    return jsonify(helm_repo_index_response['entries']), 200

@app.route('/charts/test_xapp-1.0.0.tgz', methods=['GET'])
def download_xapp_helm_package():
    if not os.path.exists(settings.MOCK_TEST_HELM_REPO_TEMP_DIR):
        os.makedirs(settings.MOCK_TEST_HELM_REPO_TEMP_DIR)
    with open(settings.MOCK_TEST_HELM_REPO_TEMP_DIR + '/values.yaml', 'w') as file:
        file.write(json.dumps(helm_repo_index_response))

    with tarfile.open(settings.MOCK_TEST_HELM_REPO_TEMP_DIR + '/test_xapp-1.0.0.tgz', "w:gz") as tar:
        tar.addfile(tarfile.TarInfo("test_xapp/values.yaml"), open(settings.MOCK_TEST_HELM_REPO_TEMP_DIR + '/values.yaml'))

    return send_file(settings.MOCK_TEST_HELM_REPO_TEMP_DIR + '/test_xapp-1.0.0.tgz', mimetype='application/zip')



@app.route('/schema.json', methods=['GET'])
def get_schema():


    return controls_schema_file, 200

@app.route('/config-file.json', methods=['GET'])
def get_config_file():


    return config_file, 200

def run():
    log.info('>>>>> Starting mock helm_repo at http://{}<<<<<'.format(listen_address))
    app.run(debug=True)

if __name__ == '__main__':
    run()






