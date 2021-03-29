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
from flask import make_response
from flask_restplus import Resource
from xapp_onboarder.api.api_reference import api
from xapp_onboarder.api.charts import get_charts_list, download_chart_package, download_values_yaml
from xapp_onboarder.api.models.response_models import error_message_model

log = logging.getLogger(__name__)
ns = api.namespace('charts', description='Managing helm charts')


@ns.route('')
class ChartsList(Resource):

    @api.response(200, 'Get helm chart list OK')
    @api.response(500, 'Get helm chart list failed', error_message_model)
    def get(self):
        """
        Returns the list of xApp helm charts that have been onboarded.
        """

        return get_charts_list()


@ns.route('/xapp/<string:xapp_chart_name>')
class VersionList(Resource):

    @api.response(200, 'Get helm chart OK')
    @api.response(500, 'Get helm chart failed', error_message_model)
    def get(self, xapp_chart_name):
        """
        Returns the helm chart for the specified xApp.
        """
        return get_charts_list(xapp_chart_name=xapp_chart_name)


@ns.route('/xapp/<string:xapp_chart_name>/ver/<string:version>')
class ChartsFetcher(Resource):

    @api.response(200, 'Get helm chart package OK')
    @api.response(500, 'Get helm chart package failed', error_message_model)
    @api.produces(['application/gzip'])
    def get(self, xapp_chart_name, version):
        """
        Returns the helm chart for the specified xApp and version.
        """

        content, status = download_chart_package(xapp_chart_name=xapp_chart_name, version=version)

        if status != 200:
            return content, status

        response = make_response(content)
        response.headers.set("Content-Type", "application/gzip")
        response.headers.set("Content-Disposition",
                             "attachment; filename=\"" + xapp_chart_name + "-" + version + ".tgz\"")
        return response


@ns.route('/xapp/<string:xapp_chart_name>/ver/<string:version>/values.yaml')
class ValuesYamlFetcher(Resource):

    @api.response(200, 'Get helm chart values.yaml OK')
    @api.response(500, 'Get helm chart values.yaml failed', error_message_model)
    @api.produces(['text/x-yaml'])
    def get(self, xapp_chart_name, version):
        """
        Returns the helm values.yaml file of the specified xApp and version.
        """

        content, status = download_values_yaml(xapp_chart_name=xapp_chart_name, version=version)

        if status != 200:
            return content, status

        response = make_response(content)
        response.headers.set("Content-Type", "text/x-yaml")
        response.headers.set("Content-Disposition", "attachment; filename=\"values.yaml\"")

        return response
