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

controls_schema_file = {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "#/controls",
    "type": "object",
    "title": "Controls Section Schema",
    "required": [
        "test"
    ],
    "properties": {
        "test": {
            "$id": "#/controls/test",
            "type": "string",
            "title": "test",
            "default": "test",
            "examples": [
                "test"
            ]
        }
    }
}

config_file = {
    "name": "test_xapp",
    "version": "1.0.0",
	"annotations": {
	  "prometheus.io/path": "/ric/v1/metrics",
      "prometheus.io/port": "8080",
      "prometheus.io/scrape": "true"
	},
    "containers": [
        {
            "name": "mcxapp",
            "image": {
                "registry": "nexus3.o-ran-sc.org:10002",
                "name": "o-ran-sc/ric-app-mc",
                "tag": "1.0.2"
            },
            "command": ["/bin/sh"],
            "args": ["-c", "/playpen/bin/container_start.sh"]
        }
    ],
    "livenessProbe": {
        "exec": {
            "command": ["/usr/local/bin/health_ck"]
        },
        "initialDelaySeconds": 5,
        "periodSeconds": 15
    },
    "readinessProbe": {
        "httpGet": {
            "path": "ric/v1/health/alive",
            "port": 8080
        },
        "initialDelaySeconds": 5,
        "periodSeconds": 15
    },
    "messaging": {
        "ports": [
            {
                "name": "http",
                "container": "mcxapp",
                "port": 8080,
                "description": "http service"
            },
            {
                "name": "rmr_data",
                "container": "mcxapp",
                "port": 4560,
                "description": "rmr data port for mcxapp"
            },
            {
                "name": "rmr_route",
                "container": "mcxapp",
                "port": 4561,
                "description": "rmr route port for mcxapp"
            }
        ],
        "maxSize": 2072,
        "numWorkers": 1,
        "txMessages": [
            "RIC_SUB_REQ",
            "RIC_SUB_DEL_REQ"
        ],
        "rxMessages": [
            "RIC_SUB_RESP",
            "RIC_SUB_FAILURE",
            "RIC_SUB_DEL_RESP",
            "RIC_INDICATION"
        ],
        "policies": [1, 2]
    },
    "controls": {
        "test": "test"
    },
    "metrics": [
        {
            "objectName": "UEEventStreamingCounters",
            "objectInstance": "SgNBAdditionRequest",
            "name": "SgNBAdditionRequest",
            "type": "counter",
            "description": "The total number of SG addition request events processed"
        },
        {
            "objectName": "UEEventStreamingCounters",
            "objectInstance": "SgNBAdditionRequestAcknowledge",
            "name": "SgNBAdditionRequestAcknowledge",
            "type": "counter",
            "description": "The total number of SG addition request acknowledge events processed"
        }
    ]
}

mock_json_body_url = {
    'config-file.json_url': 'http://0.0.0.0:8080/config-file.json',
    'controls-schema.json_url': 'http://0.0.0.0:8080/schema.json'
}

mock_json_body_url_without_controls = {
    'config-file.json_url': 'http://0.0.0.0:8080/config-file.json'
}

mock_json_body = {
    "config-file.json": config_file,
    "controls-schema.json": controls_schema_file
}

mock_json_body_without_controls = {
    "config-file.json": config_file
}
helm_repo_index_response = {'apiVersion': 'v1',
                            'entries': {
                                'test_xapp': [{
                                    'apiVersion': 'v1',
                                    'appVersion': '1.0',
                                    'created': '2020-03-12T19:10:17.178396719Z',
                                    'description': 'test xApp Helm Chart',
                                    'digest': 'd77dfb3f008e5174e90d79bfe982ef85b5dc5930141f6a1bd9995b2fa35',
                                    'name': 'test_xapp',
                                    'urls': ['charts/test-1.0.0.tgz'],
                                    'version': '1.0.0'
                                }]
                            },
                            'generated': '2020-03-16T16:54:44Z',
                            'serverInfo': {}
                            }
