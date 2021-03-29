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
import json
import copy
from jsonschema import ValidationError, SchemaError
from jsonschema import validate, Draft7Validator
from xapp_onboarder.helm_controller.xApp_builder import xApp, xAppError
from xapp_onboarder.server import settings
from xapp_onboarder.repo_manager.repo_manager import requests_retry_session, repo_manager
from xapp_onboarder.api.models.response_models import error_message_model, response, status_message_model
from xapp_onboarder.helm_controller.xapp_schema import schema as xapp_schema

log = logging.getLogger(__name__)


def onboard(config_file, controls_schema_file):
    if not repo_manager.is_repo_ready():
        response_message = response(model=error_message_model, status_code=500,
                                    error_source="xapp_onboarder",
                                    error_message="Cannot connect to local helm repo.",
                                    status="Service not ready.")
        return response_message.get_return()

    schema_file = copy.deepcopy(xapp_schema)

    if controls_schema_file:
        schema_file["properties"]["controls"] = controls_schema_file

    try:
        Draft7Validator.check_schema(schema_file)
        validate(config_file, schema_file)
    except ValidationError as err:
        log.debug(err.message)
        response_message = response(model=error_message_model, status_code=400,
                                    error_source="config-file.json",
                                    error_message=err.message,
                                    status="Input payload validation failed")
        return response_message.get_return()
    except SchemaError as err:
        log.debug(err.message)
        response_message = response(model=error_message_model, status_code=400,
                                    error_source="schema.json",
                                    error_message=err.message,
                                    status="Input payload validation failed")
        return response_message.get_return()

    try:
        xapp = xApp(config_file, schema_file)
        xapp.package_chart()
        xapp.distribute_chart()
    except xAppError as err:
        log.error(str(err))
        response_message = response(model=error_message_model, status_code=err.status_code,
                                    error_source="xApp_builder",
                                    error_message=str(err),
                                    status="xApp onboarding failed")
        return response_message.get_return()
    return response(model=status_message_model, status_code=201, status="Created").get_return()


def download_config_and_schema_and_onboard(config_file_url, controls_schema_url):
    if not repo_manager.is_repo_ready():
        response_message = response(model=error_message_model, status_code=500,
                                    error_source="xapp_onboarder",
                                    error_message="Cannot connect to local helm repo.",
                                    status="Service not ready.")
        return response_message.get_return()

    session = requests_retry_session()
    try:
        response_content = session.get(config_file_url, timeout=settings.HTTP_TIME_OUT)
    except Exception as err:
        log.error(err.message)
        response_message = response(model=error_message_model, status_code=500,
                                    error_source="config-file.json",
                                    error_message=err.message,
                                    status="Downloading config-file.json failed")
        return response_message.get_return()
    else:
        if response_content.status_code != 200:
            error_message = "Wrong response code: {}, {}".format(response_content.status_code, response_content.content.decode("utf-8"))
            log.error(error_message)
            response_message = response(model=error_message_model, status_code=500,
                                        error_source="config-file.json",
                                        error_message=error_message,
                                        status="Downloading config-file.json failed")
            return response_message.get_return()
        config_file = json.loads(response_content.content)

    controls_schema_file = None
    if controls_schema_url:
        try:
            response_content = session.get(controls_schema_url, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            log.error(err.message)
            response_message = response(model=error_message_model, status_code=500,
                                        error_source="schema.json",
                                        error_message=err.message,
                                        status="Downloading schema.json failed")
            return response_message.get_return()
        else:
            if response_content.status_code != 200:
                error_message = "Wrong response code. {}, {}".format(response_content.status_code, response_content.content.decode("utf-8"))
                log.error(error_message)
                response_message = response(model=error_message_model, status_code=500,
                                            error_source="schema.json",
                                            error_message=error_message,
                                            status="Downloading schema.json failed")
                return response_message.get_return()
            controls_schema_file = json.loads(response_content.content)


    return onboard(config_file, controls_schema_file)
