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

from http import HTTPStatus
from tests.constants import mock_json_body, mock_json_body_url_without_controls, mock_json_body_url, mock_json_body_without_controls,  helm_repo_index_response


def test_health(client):
    response = client.get('/api/v1/health')
    assert response.status_code == HTTPStatus.OK, 'Wrong status code'
    assert response.json == {'status': 'OK'}, 'Improper response'

def test_get_charts(client):
    response = client.get('/api/v1/charts')
    assert response.status_code == HTTPStatus.OK, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert sorted([repr(x) for x in response.json]) == sorted([repr(x) for x in helm_repo_index_response['entries']])

def test_get_test_xapp_charts(client):
    response = client.get('/api/v1/charts/xapp/test_xapp')
    assert response.status_code == HTTPStatus.OK, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert sorted([repr(x) for x in response.json]) == sorted([repr(x) for x in helm_repo_index_response['entries']['test_xapp']])

def test_get_test_xapp_charts_package(client):
    response = client.get('/api/v1/charts/xapp/test_xapp/ver/1.0.0')
    assert response.status_code == HTTPStatus.OK, 'Wrong status code'
    assert response.content_type == 'application/gzip', 'Content type error'

def test_get_test_xapp_charts_values_yaml(client):
    response = client.get('/api/v1/charts/xapp/test_xapp/ver/1.0.0/values.yaml')
    assert response.status_code == HTTPStatus.OK, 'Wrong status code'
    assert response.content_type == 'text/x-yaml', 'Content type error'

def test_onboard_post(client):
    url = '/api/v1/onboard'
    response = client.post(url, json=mock_json_body)
    assert response.status_code == HTTPStatus.CREATED, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert response.json == {'status': 'Created'}, 'Onboard failed'

def test_onboard_without_controls_post(client):
    url = '/api/v1/onboard'
    response = client.post(url, json=mock_json_body_without_controls)
    assert response.status_code == HTTPStatus.BAD_REQUEST, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert response.json == {'error_message': "'__empty_control_section__' is a required property",
                             'error_source': 'config-file.json',
                             'status': 'Input payload validation failed'}, 'Onboard failed'


def test_onboard_download_post(client):
    url = '/api/v1/onboard/download'
    response = client.post(url, json=mock_json_body_url)
    assert response.status_code == HTTPStatus.CREATED, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert response.json == {'status': 'Created'}, 'Onboard failed'


def test_onboard_download_without_controls_post(client):
    url = '/api/v1/onboard/download'
    response = client.post(url, json=mock_json_body_url_without_controls)
    assert response.status_code == HTTPStatus.BAD_REQUEST, 'Wrong status code'
    assert response.content_type == 'application/json', 'Content type error'
    assert response.json == {'error_message': "'__empty_control_section__' is a required property",
                             'error_source': 'config-file.json',
                             'status': 'Input payload validation failed'}, 'Onboard failed'
