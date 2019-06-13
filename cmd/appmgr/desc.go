/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type ConfigMetadata struct {
	Name       string `json:"name"`
	ConfigName string `json:"configName, omitempty"`
	Namespace  string `json:"namespace, omitempty"`
}

type XAppConfig struct {
	Metadata      ConfigMetadata `json:"metadata"`
	Descriptor    interface{}    `json:"descriptor, omitempty"`
	Configuration interface{}    `json:"config, omitempty"`
}

type ConfigMap struct {
	Kind       string      `json:"kind"`
	ApiVersion string      `json:"apiVersion"`
	Data       interface{} `json:"data"`
	Metadata   CMMetadata  `json:"metadata"`
}

type CMMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type CMError struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

func (cm *ConfigMap) UploadConfig() (cfg []XAppConfig) {
	ns := cm.getNamespace("")
	for _, name := range cm.GetNamesFromHelmRepo() {
		if name == "appmgr" {
			continue
		}

		c := XAppConfig{
			Metadata: ConfigMetadata{Name: name, Namespace: ns, ConfigName: name + "-appconfig"},
		}

		err := cm.ReadSchema(name, &c)
		if err != nil {
			continue
		}

		err = cm.ReadConfigMap(c.Metadata.ConfigName, ns, &c.Configuration)
		if err != nil {
			log.Println("No active configMap found, using default!")
		}

		cfg = append(cfg, c)
	}
	return
}

func (cm *ConfigMap) ReadSchema(name string, c *XAppConfig) (err error) {
	if err = cm.FetchChart(name); err != nil {
		return
	}

	tarDir := viper.GetString("xapp.tarDir")
	err = cm.ReadFile(path.Join(tarDir, name, viper.GetString("xapp.schema")), &c.Descriptor)
	if err != nil {
		return
	}

	err = cm.ReadFile(path.Join(tarDir, name, viper.GetString("xapp.config")), &c.Configuration)
	if err != nil {
		return
	}

	if err = os.RemoveAll(path.Join(tarDir, name)); err != nil {
		log.Println("RemoveAll failed", err)
	}

	return
}

func (cm *ConfigMap) ReadConfigMap(ConfigName string, ns string, c *interface{}) (err error) {
	args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s", ns, ConfigName)
	configMapJson, err := KubectlExec(args)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(configMapJson), &c)
	if err != nil {
		return
	}

	return
}

func (cm *ConfigMap) ApplyConfigMap(r XAppConfig, action string) (err error) {
	c := ConfigMap{
		Kind:       "ConfigMap",
		ApiVersion: "v1",
		Metadata:   CMMetadata{Name: r.Metadata.Name, Namespace: r.Metadata.Namespace},
		Data:       r.Configuration,
	}

	cmJson, err := json.Marshal(c.Data)
	if err != nil {
		log.Println("Config marshalling failed: ", err)
		return
	}

	cmFile := viper.GetString("xapp.tmpConfig")
	err = ioutil.WriteFile(cmFile, cmJson, 0644)
	if err != nil {
		log.Println("WriteFile failed: ", err)
		return
	}

	cmd := " create configmap -n %s %s --from-file=%s -o json --dry-run | kubectl %s -f -"
	args := fmt.Sprintf(cmd, r.Metadata.Namespace, r.Metadata.ConfigName, cmFile, action)
	_, err = KubectlExec(args)
	if err != nil {
		return
	}
	log.Println("Configmap changes done!")

	return
}

func (cm *ConfigMap) GetConfigMap(m XappDeploy, c *interface{}) (err error) {
	if m.ConfigName == "" {
		m.ConfigName = m.Name + "-appconfig"
	}
	return cm.ReadConfigMap(m.ConfigName, m.Namespace, c)
}

func (cm *ConfigMap) CreateConfigMap(r XAppConfig) (errList []CMError, err error) {
	if errList, err = cm.Validate(r); err != nil {
		return
	}
	err = cm.ApplyConfigMap(r, "create")
	return
}

func (cm *ConfigMap) UpdateConfigMap(r XAppConfig) (errList []CMError, err error) {
	if errList, err = cm.Validate(r); err != nil {
		return
	}

	// Re-create the configmap with the new parameters
	err = cm.ApplyConfigMap(r, "apply")
	return
}

func (cm *ConfigMap) DeleteConfigMap(r XAppConfig) (c interface{}, err error) {
	err = cm.ReadConfigMap(r.Metadata.ConfigName, r.Metadata.Namespace, &c)
	if err == nil {
		args := fmt.Sprintf(" delete configmap --namespace=%s %s", r.Metadata.Namespace, r.Metadata.ConfigName)
		_, err = KubectlExec(args)
	}
	return
}

func (cm *ConfigMap) PurgeConfigMap(m XappDeploy) (c interface{}, err error) {
	if m.ConfigName == "" {
		m.ConfigName = m.Name + "-appconfig"
	}
	md := ConfigMetadata{Name: m.Name, Namespace: m.Namespace, ConfigName: m.ConfigName}

	return cm.DeleteConfigMap(XAppConfig{Metadata: md})
}

func (cm *ConfigMap) RestoreConfigMap(m XappDeploy, c interface{}) (err error) {
	if m.ConfigName == "" {
		m.ConfigName = m.Name + "-appconfig"
	}
	md := ConfigMetadata{Name: m.Name, Namespace: m.Namespace, ConfigName: m.ConfigName}
	time.Sleep(time.Duration(10 * time.Second))

	return cm.ApplyConfigMap(XAppConfig{Metadata: md, Configuration: c}, "create")
}

func (cm *ConfigMap) GetNamesFromHelmRepo() (names []string) {
	rname := viper.GetString("helm.repo-name")

	cmdArgs := strings.Join([]string{"search ", rname}, "")
	out, err := HelmExec(cmdArgs)
	if err != nil {
		return
	}

	re := regexp.MustCompile(rname + `/.*`)
	result := re.FindAllStringSubmatch(string(out), -1)
	if result != nil {
		var tmp string
		for _, v := range result {
			fmt.Sscanf(v[0], "%s", &tmp)
			names = append(names, strings.Split(tmp, "/")[1])
		}
	}
	return names
}

func (cm *ConfigMap) Validate(req XAppConfig) (errList []CMError, err error) {
	c := XAppConfig{}
	err = cm.ReadSchema(req.Metadata.Name, &c)
	if err != nil {
		log.Printf("No schema file found for '%s', aborting ...", req.Metadata.Name)
		return
	}
	return cm.doValidate(c.Descriptor, req.Configuration)
}

func (cm *ConfigMap) doValidate(schema, cfg interface{}) (errList []CMError, err error) {
	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(cfg)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Println("Validation failed: ", err)
		return
	}

	if result.Valid() == false {
		log.Println("The document is not valid, Errors: ", result.Errors())
		for _, desc := range result.Errors() {
			errList = append(errList, CMError{Field: desc.Field(), Description: desc.Description()})
		}
		return errList, errors.New("Validation failed!")
	}
	return
}

func (cm *ConfigMap) ReadFile(name string, data interface{}) (err error) {
	f, err := ioutil.ReadFile(name)
	if err != nil {
		log.Printf("Reading '%s' file failed: %v", name, err)
		return
	}

	err = json.Unmarshal(f, &data)
	if err != nil {
		log.Printf("Unmarshalling '%s' file failed: %v", name, err)
		return
	}

	return
}

func (cm *ConfigMap) FetchChart(name string) (err error) {
	tarDir := viper.GetString("xapp.tarDir")
	repo := viper.GetString("helm.repo-name")
	fetchArgs := fmt.Sprintf("--untar --untardir %s %s/%s", tarDir, repo, name)

	_, err = HelmExec(strings.Join([]string{"fetch ", fetchArgs}, ""))
	return
}

func (cm *ConfigMap) GetMessages(name string) (msgs MessageTypes) {
	log.Println("Fetching tx/rx messages for: ", name)

	ns := cm.getNamespace("")
	args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s-appconfig", ns, name)
	out, err := KubectlExec(args)
	if err != nil {
		return
	}

	var p fastjson.Parser
	v, err := p.Parse(string(out))
	if err != nil {
		log.Printf("fastjson.Parser for '%s' failed: %v", name, err)
		return
	}

	for _, m := range v.GetArray("rmr", "txMessages") {
		msgs.TxMessages = append(msgs.TxMessages, strings.Trim(m.String(), `"`))
	}
	for _, m := range v.GetArray("rmr", "rxMessages") {
		msgs.RxMessages = append(msgs.RxMessages, strings.Trim(m.String(), `"`))
	}

	return
}

func (cm *ConfigMap) getNamespace(ns string) string {
	if ns != "" {
		return ns
	}

	ns = viper.GetString("xapp.namespace")
	if ns == "" {
		ns = "ricxapp"
	}
	return ns
}
