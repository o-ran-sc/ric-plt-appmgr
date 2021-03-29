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
from xapp_onboarder.server import settings
import logging
from flask import Flask, Blueprint
from xapp_onboarder.api.api_reference import api
from xapp_onboarder.api.endpoints.onboard_ep import ns as onboard_ns
from xapp_onboarder.api.endpoints.health_check_ep import ns as health_check_ns
from xapp_onboarder.api.endpoints.charts_ep import ns as charts_ns
from xapp_onboarder.helm_controller.artifacts_manager import artifacts_manager

log = logging.getLogger(__name__)

class server(object):
    def __init__(self):
        self.app = Flask(__name__)
        self.app.config['SWAGGER_UI_DOC_EXPANSION'] = settings.RESTPLUS_SWAGGER_UI_DOC_EXPANSION
        self.app.config['RESTPLUS_VALIDATE'] = settings.RESTPLUS_VALIDATE
        self.app.config['RESTPLUS_MASK_SWAGGER'] = settings.RESTPLUS_MASK_SWAGGER
        self.app.config['ERROR_404_HELP'] = settings.RESTPLUS_ERROR_404_HELP
        self.api = api
        self.api_bp = Blueprint('api', __name__, url_prefix='/api/v1')
        self.api.init_app(self.api_bp)
        self.api.add_namespace(onboard_ns)
        self.api.add_namespace(health_check_ns)
        self.api.add_namespace(charts_ns)
        self.app.register_blueprint(self.api_bp)
        self.artifacts_manager = artifacts_manager()
        self.artifacts_manager.start()

    def run(self):
        log.info('>>>>> Starting development xapp_onboarder at http://{}/api/v1/ <<<<<'.format(self.app.config['SERVER_NAME']))
        self.app.run(debug=settings.FLASK_DEBUG, host='0.0.0.0', port=settings.FLASK_PORT)





def main():
    logger_config = pkg_resources.resource_filename("xapp_onboarder", 'logging.conf')
    logging.config.fileConfig(logger_config)
    server().run()

if __name__ == "__main__":
    main()

