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

import yaml
import json
from xapp_onboarder.server import settings
import logging
import requests
import time
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry

log = logging.getLogger(__name__)

def requests_retry_session(retries=3, backoff_factor=0.3, status_forcelist=(500, 502, 504, 400, 401, 409), session=None,):
    session = session or requests.Session()
    retry = Retry(
        total=retries,
        read=retries,
        connect=retries,
        backoff_factor=backoff_factor,
        status_forcelist=status_forcelist,
    )
    adapter = HTTPAdapter(max_retries=retry)
    session.mount('http://', adapter)
    session.mount('https://', adapter)
    return session



class RepoManagerError(Exception):
    def __init__(self, message, status_code):
        # Call the base class constructor with the parameters it needs
        super().__init__(message)
        self.status_code = status_code


class repoManager():
    def __init__(self, repo_url):
        self.repo_url = repo_url
        self.__is_repo_ready__ = False
        log.debug("Initialize connection to helm chart repo at "+self.repo_url)
        t0 = time.time()
        self.retry_session = requests_retry_session()
        try:
            response = self.retry_session.get(self.repo_url, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            t1 = time.time()
            log.error('Failed to connect to helm chart repo ' + self.repo_url + ' after ' + str(
                settings.HTTP_RETRY) + ' retries and ' + str(t1 - t0) + ' seconds. (Caused by: ' + err.__class__.__name__ + ')')
        else:
            self.__is_repo_ready__ = True



    def is_repo_ready(self):
        return self.__is_repo_ready__

    def get_index(self):
        try:
            response = self.retry_session.get(self.repo_url +'/index.yaml', timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            raise RepoManagerError("Get helm repo index failed. (Caused by: " + str(err)+")", 500)
        else:
            if response.status_code != 200:
                raise RepoManagerError("Get helm repo index failed. Helm repo return status code: {}, {}".format(response.status_code, response.content.decode("utf-8")))
            return yaml.load(response.content, Loader=yaml.FullLoader)

    def upload_chart(self, xapp):

        xapp_chart_index = self.get_index()
        found_xapp = False
        for chart in xapp_chart_index.get('entries', {}).get(xapp.chart_name, []):
            if chart['version'] == xapp.chart_version:
                found_xapp = True

        if found_xapp:
            if settings.ALLOW_REDEPLOY:
                self.delete_chart(xapp)
            else:
                raise RepoManagerError("Upload helm chart failed. Redeploy xApp helm chart is not allowed.", 400)

        headers = {'Content-Type': 'application/json'}
        chart_package_path = xapp.chart_workspace_path + '/' + xapp.chart_name + '-' + xapp.chart_version + '.tgz'
        with open(chart_package_path, mode='rb') as filereader:
            fileContent = filereader.read()

        try:
            response = self.retry_session.post(self.repo_url +'/api/charts', headers=headers, data=fileContent, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            raise RepoManagerError("Upload helm chart failed. (Caused by: " + str(err) + ")", 500)
        else:
            if response.status_code != 201:
                raise RepoManagerError("Upload helm chart failed. Helm repo return status code: "+ str(response.status_code)  +" "+ response.content.decode("utf-8"), response.status_code)


    def delete_chart(self, xapp):

        headers = {'Content-Type': 'application/json'}

        try:
            response = self.retry_session.delete(self.repo_url +'/api/charts/' + xapp.chart_name
                                                 + '/' + xapp.chart_version, headers=headers, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            raise RepoManagerError("Delete helm chart failed." + str(err), 500)
        else:
            if response.status_code != 200:
                response_dict = json.loads(response.content)
                if xapp.chart_name+'-'+xapp.chart_version+'.tgz' not in response_dict["error"]:
                    raise RepoManagerError("Delete helm chart failed. Helm repo return status code:" + str(response.status_code)  +" "+ response.content.decode("utf-8"),response.status_code)


    def get_xapp_list(self, xapp_chart_name=None):

        request_path = self.repo_url+'/api/charts'
        if xapp_chart_name:
            request_path = request_path +'/' + xapp_chart_name

        try:
            response = self.retry_session.get(request_path, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            raise RepoManagerError("Get xApp charts list failed. (Caused by: " + str(err)+")", 500)
        else:
            if response.status_code != 200:
                raise RepoManagerError("Get xApp charts list failed. Helm repo return status code: "+ str(response.status_code)  +" "+ response.content.decode("utf-8"), response.status_code)
            return json.loads(response.content)


    def download_xapp_chart(self, xapp_chart_name, version):

        request_path = self.repo_url+'/charts/'+xapp_chart_name+'-'+version+'.tgz'
        try:
            response = self.retry_session.get(request_path, timeout=settings.HTTP_TIME_OUT)
        except Exception as err:
            raise RepoManagerError("Download helm chart failed. (Caused by: " + str(err)+")", 500)
        else:
            if response.status_code != 200:
                raise RepoManagerError( "Download helm chart failed. Helm repo return status code: "+ str(response.status_code)  +" "+ response.content.decode("utf-8"), response.status_code)
            return response.content




repo_manager = repoManager(settings.CHART_REPO_URL)
