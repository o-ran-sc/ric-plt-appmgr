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

import os

# Flask settings
FLASK_PORT = os.environ.get('FLASK_PORT') or '8888'
FLASK_DEBUG = os.environ.get('FLASK_DEBUG') or True  # Do not use debug mode in production

# Flask-Restplus settings
RESTPLUS_SWAGGER_UI_DOC_EXPANSION = os.environ.get('RESTPLUS_SWAGGER_UI_DOC_EXPANSION') or 'list'
RESTPLUS_VALIDATE = os.environ.get('RESTPLUS_VALIDATE') or True
RESTPLUS_MASK_SWAGGER = os.environ.get('RESTPLUS_MASK_SWAGGER') or False
RESTPLUS_ERROR_404_HELP = os.environ.get('RESTPLUS_ERROR_404_HELP') or False

# xapp_onboarder settings
CHART_WORKSPACE_PATH = os.environ.get('CHART_WORKSPACE_PATH') or '/tmp/xapp_onboarder'
CHART_REPO_URL = os.environ.get('CHART_REPO_URL') or 'http://0.0.0.0:8080'
HTTP_TIME_OUT = int(os.environ.get('HTTP_TIME_OUT') or 10)
HELM_VERSION = os.environ.get('HELM_VERSION') or '2.12.3'
HTTP_RETRY = os.environ.get('HTTP_RETRY') or 3
ALLOW_REDEPLOY = os.environ.get('ALLOW_REDEPLOY') or True
CHART_WORKSPACE_SIZE = os.environ.get('CHART_WORKSPACE_SIZE') or '500 MB'
MOCK_TEST_MODE = os.environ.get('MOCK_TEST_MODE') or False
MOCK_TEST_HELM_REPO_TEMP_DIR = os.environ.get('MOCK_TEST_HELM_REPO_TEMP_DIR') or '/tmp/mock_helm_repo'


