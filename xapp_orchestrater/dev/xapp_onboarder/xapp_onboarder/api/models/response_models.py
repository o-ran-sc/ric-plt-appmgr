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

from flask_restplus import fields, marshal
from xapp_onboarder.api.api_reference import api

error_message_model = api.model('error_message', {
    'error_source': fields.String(description='source of the error', required=True),
    'error_message': fields.String(description='source of the error', required=True),
    'status': fields.String(description='http response message', required=True),
})

status_message_model = api.model('status', {
    'status': fields.String(description='status of the service', required=True)
})


class response():

    def __init__(self, model, status_code, status = "" , error_source = "", error_message = ""):
        self.model = model
        self.status = status
        self.status_code = status_code
        self.error_source = error_source
        self.error_message = error_message

    def get_return(self):
        return marshal(self, self.model), self.status_code