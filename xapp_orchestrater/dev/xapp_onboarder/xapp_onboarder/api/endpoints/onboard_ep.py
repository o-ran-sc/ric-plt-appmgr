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

import logging
from flask import request
from flask_restplus import Resource
from xapp_onboarder.api.models import request_models
from xapp_onboarder.api.api_reference import api
from xapp_onboarder.api.models.response_models import status_message_model, error_message_model
from xapp_onboarder.api.onboard import onboard, download_config_and_schema_and_onboard

log = logging.getLogger(__name__)
ns = api.namespace('onboard', description='onboard xApps')


@ns.route('')
class OnboardxApps(Resource):

    # @api.response(200, 'Everything is fine')
    # @api.response(500, 'temp is not ready')
    # def get(self):
    #     """
    #     Return a list of xApp that have been onboarded and their versions.
    #     """
    #     if not repo_manager.is_repo_ready():
    #         return {'status': 'not ready'}, 500
    #     return {'status': 'OK'}, 200

    @api.response(201, 'xApp onboard successfully.', status_message_model)
    @api.response(400, 'xApp descriptor format error', error_message_model)
    @api.response(500, 'xApp onboarder is not ready', error_message_model)
    @api.expect(request_models.xapp_descriptor_post, validate=True)
    def post(self):
        """
        Onboard xApp using the xApp descriptor and schema in the request body.
        """
        config_file = request.json.get('config-file.json')
        controls_schema_file = request.json.get('controls-schema.json')

        return onboard(config_file, controls_schema_file)


@ns.route('/download')
class OnboardxAppsDownload(Resource):

    @api.response(201, 'xApp onboard successfully.', status_message_model)
    @api.response(400, 'xApp descriptor format error', error_message_model)
    @api.response(500, 'xApp onboarder is not ready', error_message_model)
    @api.expect(request_models.xapp_descriptor_download_post, validate=True)
    def post(self):
        """
        Onboard xApp after downloading the xApp descriptor and schema from the URLs.
        """
        config_file_url = request.json.get('config-file.json_url')
        controls_schema_url = request.json.get('controls-schema.json_url')

        return download_config_and_schema_and_onboard(config_file_url, controls_schema_url)
