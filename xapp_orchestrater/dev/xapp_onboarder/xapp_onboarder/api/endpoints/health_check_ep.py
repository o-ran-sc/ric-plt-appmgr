###############################################################################
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

from flask_restplus import Resource
from xapp_onboarder.api.api_reference import api
from xapp_onboarder.api.models.response_models import status_message_model, error_message_model, response
from xapp_onboarder.repo_manager.repo_manager import repo_manager

log = logging.getLogger(__name__)
ns = api.namespace('health', description='health check')


@ns.route('')
class HealthCheck(Resource):

    @api.response(200, 'Health check OK', status_message_model)
    @api.response(500, 'xApp onboarder is not ready', error_message_model)
    def get(self):
        """
        Returns the health condition of the xApp onboarder
        """
        if not repo_manager.is_repo_ready():
            response_message = response( model= error_message_model, status_code = 500,
                                         error_source = "xapp_onboarder",
                                         error_message = "Cannot connect to local helm repo.",
                                         status = "Service not ready.")
            return response_message.get_return()
        return response(model = status_message_model, status_code = 200, status= "OK").get_return()
