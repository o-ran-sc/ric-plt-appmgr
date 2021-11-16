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
import yaml
import json
import os
import io
import subprocess
import shutil
import re
import copy
import platform
import tarfile
import stat
import sys
from xapp_onboarder.server import settings
from xapp_onboarder.repo_manager.repo_manager import repo_manager, RepoManagerError
from pkg_resources import resource_filename
from subprocess import PIPE, check_output, STDOUT
from xapp_onboarder.repo_manager.repo_manager import requests_retry_session
log = logging.getLogger(__name__)


def indent(text, amount, ch=' '):
    padding = amount * ch
    return ''.join(padding + line for line in text.splitlines(True))


class xAppError(Exception):
    def __init__(self, message, status_code):
        # Call the base class constructor with the parameters it needs
        super().__init__(message)
        self.status_code = status_code


class xApp():
    def __init__(self, config_file, schema_file):
        self.config_file = config_file
        self.schema_file = schema_file
 
        isnamepresent = 0
        if 'name' not in self.config_file:
            isnamepresent = 1
            if 'xapp_name' not in self.config_file:
                raise xAppError(
                    "xApp chart name not found. (Caused by: config-file.json does not contain xapp_name attribute.)", 500)

        if 'version' not in self.config_file:
            raise xAppError(
                "xApp chart version not found. (Caused by: config-file.json does not contain version attribute.)", 500)
        
        if isnamepresent == 1:
            self.chart_name = self.config_file['xapp_name']
        else:
            self.chart_name = self.config_file['name']
        self.chart_version = self.config_file['version']
        self.configmap_config_json_file = copy.deepcopy(self.config_file)
        self.chart_workspace_path = settings.CHART_WORKSPACE_PATH + '/' + self.chart_name + '-' + self.chart_version
        if os.path.exists(self.chart_workspace_path):
            shutil.rmtree(self.chart_workspace_path)
        os.makedirs(self.chart_workspace_path)
        shutil.copytree(resource_filename( 'xapp_onboarder', 'resources/xapp-std'), self.chart_workspace_path + '/' + self.chart_name)
        self.setup_helm()

    def setup_helm(self):
        self.helm_client_path = 'helm'
        try:
            process = subprocess.run([self.helm_client_path], stdout=PIPE, stderr=PIPE, check=True)

        except Exception as err:
            print(err)
            self.download_helm()
            self.helm_client_path = settings.CHART_WORKSPACE_PATH + '/helm'

    def download_helm(self):
        if not os.path.isfile(settings.CHART_WORKSPACE_PATH + '/helm'):
            log.info("Helm client missing. Trying to download it.")
            helm_file_name = "helm-v{}-{}-amd64.tar.gz".format(settings.HELM_VERSION, platform.system().lower())
            helm_download_link = "https://get.helm.sh/" + helm_file_name


            try:
                response = requests_retry_session().get(helm_download_link, timeout=settings.HTTP_TIME_OUT)
            except Exception as err:
                error_message = "Download helm client failed. (Caused by: " + str(err)+")"
                log.error(error_message)
                raise xAppError(error_message, 500)
            else:
                if response.status_code != 200:
                    error_message = "Download helm chart failed. Helm repo return status code: "+ str(response.status_code)  +" "+ response.content.decode("utf-8")
                    log.error(error_message)
                    raise xAppError(error_message, 500)

                file_stream = io.BytesIO(response.content)

                with tarfile.open(fileobj=file_stream) as tar:
                    helm_client = tar.extractfile(platform.system().lower() + "-amd64/helm")
                    with open(settings.CHART_WORKSPACE_PATH+'/helm', 'wb') as file:
                        file.write(helm_client.read())
                st = os.stat(settings.CHART_WORKSPACE_PATH+'/helm')
                os.chmod(settings.CHART_WORKSPACE_PATH+'/helm', st.st_mode | stat.S_IEXEC)




    def recursive_convert_config_file(self, node_list=list()):
        current_node = self.configmap_config_json_file
        helm_value_path = '.Values'
        for node in node_list:
            current_node = current_node.get(node)
            helm_value_path = helm_value_path + ' ' + "\"" + node + "\""

        if type(current_node) is not dict:
            raise TypeError("Recursive write was called on a leaf node.")

        for item in current_node.keys():
            if type(current_node.get(item)) is not dict:
                current_node[item] = '{{ index '+ helm_value_path +' "'+ item + '" | toJson }}'
            else:
                new_node_list = node_list.copy()
                new_node_list.append(item)
                self.recursive_convert_config_file(new_node_list)


    def append_config_to_config_map(self):
        with open(self.chart_workspace_path + '/' + self.chart_name + '/templates/appconfig.yaml', 'a') as outputfile:
            self.recursive_convert_config_file()
            config_file_json_text = json.dumps(self.configmap_config_json_file, indent=4)
            indented_config_text = indent(config_file_json_text, 4)
            indented_config_text = re.sub(r"\"{{", '{{', indented_config_text)
            indented_config_text = re.sub(r"}}\"", '}}', indented_config_text)
            indented_config_text = re.sub(r"\\", '', indented_config_text)
            outputfile.write("  config-file.json: |\n")
            outputfile.write(indented_config_text)
            outputfile.write("\n  schema.json: |\n")
            schema_json = json.dumps(self.schema_file, indent=4)
            indented_schema_text = indent(schema_json, 4)
            outputfile.write(indented_schema_text)


# This is a work around for the bronze release to be backward compatible to the previous xapp standard helm template
    def write_config_and_schema(self):
        os.makedirs(self.chart_workspace_path + '/' + self.chart_name + '/descriptors')
        os.makedirs(self.chart_workspace_path + '/' + self.chart_name + '/config')
        with open(self.chart_workspace_path + '/' + self.chart_name + '/descriptors/schema.json', 'w') as outfile:
            json.dump(self.schema_file, outfile)
        with open(self.chart_workspace_path + '/' + self.chart_name + '/config/config-file.json', 'w') as outfile:
            json.dump(self.config_file, outfile)



    def add_probes_to_deployment(self):
        with open(self.chart_workspace_path + '/' + self.chart_name + '/templates/deployment.yaml', 'a') as outputfile:

            for probes in ['readinessProbe', 'livenessProbe']:
                if self.configmap_config_json_file.get(probes):
                    probe_definition = self.configmap_config_json_file.get(probes)
                    probe_definition_yaml = yaml.dump(probe_definition, width=1000)

                    print(probe_definition_yaml)

                    indented_probe_definition_yaml = indent(probe_definition_yaml, 12)
                    indented_probe_definition_yaml = re.sub(r" \| toJson", '', indented_probe_definition_yaml)
                    indented_probe_definition_yaml = re.sub(r"'", '', indented_probe_definition_yaml)
                    outputfile.write("          "+probes+":\n")
                    outputfile.write(indented_probe_definition_yaml)


    def append_config_to_values_yaml(self):
        with open(self.chart_workspace_path + '/' + self.chart_name + '/values.yaml', 'a') as outputfile:
            yaml.dump(self.config_file, outputfile, default_flow_style=False)


    def change_chart_name_version(self):
        with open(self.chart_workspace_path + '/' + self.chart_name + '/Chart.yaml', 'r') as inputfile:
            self.chart_yaml = yaml.load(inputfile, Loader=yaml.FullLoader)
            self.chart_yaml['version'] = self.chart_version
            self.chart_yaml['name'] = self.chart_name

        with open(self.chart_workspace_path + '/' + self.chart_name + '/Chart.yaml', 'w') as outputfile:
            yaml.dump(self.chart_yaml, outputfile, default_flow_style=False)


    def helm_lint(self):
        try:
            process = subprocess.run([self.helm_client_path, "lint", self.chart_workspace_path + "/" + self.chart_name], stdout=PIPE, stderr=PIPE, check=True)

        except OSError as err:
            raise xAppError(
                "xApp " + self.chart_name + '-' + self.chart_version + " helm lint failed. (Caused by: " + str(
                    err) + ")", 500)
        except subprocess.CalledProcessError as err:
            raise xAppError(
                "xApp " + self.chart_name + '-' + self.chart_version + " helm lint failed. (Caused by: " +
                err.stderr.decode("utf-8") +  "\n" + err.stdout.decode("utf-8") + ")", 400)

    def package_chart(self):
        self.write_config_and_schema()
        self.append_config_to_config_map()
        self.append_config_to_values_yaml()
        self.add_probes_to_deployment()
        self.change_chart_name_version()
        self.helm_lint()
        try:
            process = subprocess.run([self.helm_client_path, "package", self.chart_workspace_path + "/" + self.chart_name, "-d"
                               ,self.chart_workspace_path], stdout=PIPE, stderr=PIPE, check=True)

        except OSError as err:
                raise xAppError("xApp "+ self.chart_name+'-'+self.chart_version +" packaging failed. (Caused by: "+str(err) +")", 500)
        except subprocess.CalledProcessError as err:
            raise xAppError(
                "xApp " + self.chart_name + '-' + self.chart_version + " packaging failed. (Caused by: " +
                    err.stderr.decode("utf-8") + ")", 500)



    def distribute_chart(self):
        try:
            repo_manager.upload_chart(self)
        except RepoManagerError as err:
            raise xAppError( "xApp " + self.chart_name + '-' + self.chart_version + " distribution failed. (Caused by: " + str(err) + ")" , err.status_code)

    def install_chart_package(xapp_chart_name, version, namespace, overridefile):
        dirTomove = "/tmp/helm_template"
        try: 
          tar = tarfile.open(xapp_chart_name + "-" + version + ".tgz")
          tar.extractall()
          tar.close()
          if overridefile != "":
            process = subprocess.run(["helm", "install", xapp_chart_name, "./" + xapp_chart_name, "-f", overridefile, "--namespace=" + namespace], stdout=PIPE, stderr=PIPE, check=True)
          else:
            process = subprocess.run(["helm", "install", xapp_chart_name, "./" + xapp_chart_name, "--namespace=" + namespace], stdout=PIPE, stderr=PIPE, check=True)
          status = 1
        except subprocess.CalledProcessError as err:
            print(err.stderr.decode())
            status=0
        except Exception as err:
            print(err)
            status = 0
        if (os.getcwd() != dirTomove):
            subprocess.run(["mv", xapp_chart_name, dirTomove])
            PATH=xapp_chart_name + "-" + version + ".tgz"
            if os.path.isfile(PATH):
                subprocess.run(["mv", xapp_chart_name + "-" + version + ".tgz", dirTomove ])
        return status

    def uninstall_chart_package(xapp_chart_name, namespace, version):
        dirTomove = "/tmp/helm_template/"
        try:
          subprocess.run(["rm", "-rf", dirTomove + xapp_chart_name])
          if version != "" :
            subprocess.run(["rm", "-rf", dirTomove+xapp_chart_name + "-" + version + ".tgz"])
          process = subprocess.run(["helm", "delete", xapp_chart_name, "--namespace=" + namespace], stdout=PIPE, stderr=PIPE, check=True)
          status = 1

        except Exception as err:
                print(err.stderr.decode())
                status = 0

        return status
    def health_check_xapp(xapp_chart_name, namespace):
       
        try:
          getpodname=subprocess.check_output("kubectl get po -n " + namespace + " |  grep -w " +  xapp_chart_name + " | awk '{print $1}'", shell=True).decode().strip("\n")
          if getpodname=="":
              print("No " + xapp_chart_name + " xapp found under " + namespace + " namespace.")
              sys.exit()
          process = subprocess.check_output("kubectl describe po " + getpodname +  " --namespace=" + namespace + "| grep -B 0 -A 5 'Conditions:'", shell=True).decode()

          final= re.search("Initialized.*", process)
          temp=final.group().split(' ',1)[1]
          initialized=" ".join(temp.split())
          
          final= re.search("Ready.*", process)
          temp=final.group().split(' ',1)[1]
          ready=" ".join(temp.split())
          
          final= re.search("ContainersReady.*", process)
          temp=final.group().split(' ',1)[1]
          containersready=" ".join(temp.split())
          
          final= re.search("PodScheduled.*", process)
          temp=final.group().split(' ',1)[1]
          podscheduled=" ".join(temp.split())
	  
          if "True"==initialized and "True"==podscheduled and "True"==containersready and "True"==ready:
             print("Xapp health status : Healthy")
          else:
             print("Xapp health status : Unhealthy")
             if "True"!=containersready:
               print("ContainersReady=False, All the containers in the pod are not ready\n")
             elif "True"!=initialized:
               print("Initialized=False, Init containers have not yet started\n")
             elif "True"!=podscheduled:
               print("PodScheduled=False, Pod has not yet scheduled to node\n")
             elif "True"!=ready:
               print("Ready=False, Pod is not ready to serve any request\n")
        except Exception as err:
            print(err.output.decode())
