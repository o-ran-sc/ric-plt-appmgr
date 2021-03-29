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
import shutil
from xapp_onboarder.helm_controller.xApp_builder import xApp, xAppError
from tests.constants import config_file, controls_schema_file
from xapp_onboarder.server import settings


def test_packaging_xapp(client):
        chart_workspace_path = settings.CHART_WORKSPACE_PATH + '/test_xapp-1.0.0'
        if os.path.exists(chart_workspace_path):
                shutil.rmtree(chart_workspace_path)

        xapp = xApp(config_file, controls_schema_file)
        xapp.package_chart()
        assert os.path.isfile(chart_workspace_path + '/test_xapp-1.0.0.tgz'), 'xApp packaging error'


