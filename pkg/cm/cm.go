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

package cm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/util"
)

type CM struct{}

func NewCM() *CM {
	return &CM{}
}

func (cm *CM) UploadConfig() (cfg models.AllXappConfig) {
	ns := cm.GetNamespace("")
	for _, name := range cm.GetNamesFromHelmRepo() {
		if name == "appmgr" {
			continue
		}

		c := models.XAppConfig{
			Metadata: &models.ConfigMetadata{Name: &name, Namespace: ns, ConfigName: cm.GetConfigMapName(name, ns)},
		}

		err := cm.ReadSchema(name, &c)
		if err != nil {
			continue
		}

		err = cm.ReadConfigMap(c.Metadata.ConfigName, ns, &c.Config)
		if err != nil {
			appmgr.Logger.Info("No active configMap found, using default!")
		}

		cfg = append(cfg, &c)
	}
	return
}

func (cm *CM) ReadSchema(name string, c *models.XAppConfig) (err error) {
	if err = cm.FetchChart(name); err != nil {
		return
	}

	tarDir := viper.GetString("xapp.tarDir")
	err = cm.ReadFile(path.Join(tarDir, name, viper.GetString("xapp.schema")), &c.Descriptor)
	if err != nil {
		return
	}

	err = cm.ReadFile(path.Join(tarDir, name, viper.GetString("xapp.config")), &c.Config)
	if err != nil {
		return
	}

	if err = os.RemoveAll(path.Join(tarDir, name)); err != nil {
		appmgr.Logger.Info("RemoveAll failed: %v", err)
	}

	return
}

func (cm *CM) ReadConfigMap(ConfigName string, ns string, c *interface{}) (err error) {
	args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s", ns, ConfigName)
	configMapJson, err := util.KubectlExec(args)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(configMapJson), &c)
	if err != nil {
		return
	}

	return
}

func (cm *CM) ApplyConfigMap(r models.XAppConfig, action string) (err error) {
	c := appmgr.ConfigMap{
		Kind:       "ConfigMap",
		ApiVersion: "v1",
		Metadata:   appmgr.CMMetadata{Name: *r.Metadata.Name, Namespace: r.Metadata.Namespace},
		Data:       r.Config,
	}

	cmJson, err := json.Marshal(c.Data)
	if err != nil {
		appmgr.Logger.Info("Config marshalling failed: %v", err)
		return
	}

	cmFile := viper.GetString("xapp.tmpConfig")
	err = ioutil.WriteFile(cmFile, cmJson, 0644)
	if err != nil {
		appmgr.Logger.Info("WriteFile failed: %v", err)
		return
	}

	cmd := " create configmap -n %s %s --from-file=%s -o json --dry-run | kubectl %s -f -"
	args := fmt.Sprintf(cmd, r.Metadata.Namespace, r.Metadata.ConfigName, cmFile, action)
	_, err = util.KubectlExec(args)
	if err != nil {
		return
	}
	appmgr.Logger.Info("Configmap changes done!")

	return
}

func (cm *CM) GetConfigMap(m models.XappDescriptor, c *interface{}) (err error) {
	return cm.ReadConfigMap(cm.GetConfigMapName(*m.XappName, m.Namespace), m.Namespace, c)
}

func (cm *CM) CreateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	if errList, err = cm.Validate(r); err != nil {
		return
	}
	err = cm.ApplyConfigMap(r, "create")
	return
}

func (cm *CM) UpdateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	if errList, err = cm.Validate(r); err != nil {
		return
	}

	// Re-create the configmap with the new parameters
	err = cm.ApplyConfigMap(r, "apply")
	return
}

func (cm *CM) DeleteConfigMap(r models.ConfigMetadata) (c interface{}, err error) {
	err = cm.ReadConfigMap(r.ConfigName, r.Namespace, &c)
	if err == nil {
		args := fmt.Sprintf(" delete configmap --namespace=%s %s", r.Namespace, r.ConfigName)
		_, err = util.KubectlExec(args)
	}
	return
}

func (cm *CM) PurgeConfigMap(m models.XappDescriptor) (c interface{}, err error) {
	md := models.ConfigMetadata{Name: m.XappName, Namespace: m.Namespace, ConfigName: cm.GetConfigMapName(*m.XappName, m.Namespace)}

	return cm.DeleteConfigMap(md)
}

func (cm *CM) RestoreConfigMap(m models.XappDescriptor, c interface{}) (err error) {
	md := &models.ConfigMetadata{Name: m.XappName, Namespace: m.Namespace, ConfigName: cm.GetConfigMapName(*m.XappName, m.Namespace)}
	time.Sleep(time.Duration(10 * time.Second))

	return cm.ApplyConfigMap(models.XAppConfig{Metadata: md, Config: c}, "create")
}

func (cm *CM) GetNamesFromHelmRepo() (names []string) {
	rname := viper.GetString("helm.repo-name")

	cmdArgs := strings.Join([]string{"search ", rname}, "")
	out, err := util.HelmExec(cmdArgs)
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

func (cm *CM) Validate(req models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	c := models.XAppConfig{}
	err = cm.ReadSchema(*req.Metadata.Name, &c)
	if err != nil {
		appmgr.Logger.Info("No schema file found for '%s', aborting ...", *req.Metadata.Name)
		return
	}
	return cm.doValidate(c.Descriptor, req.Config)
}

func (cm *CM) doValidate(schema, cfg interface{}) (errList models.ConfigValidationErrors, err error) {
	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(cfg)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		appmgr.Logger.Info("Validation failed: %v", err)
		return
	}

	if result.Valid() == false {
		appmgr.Logger.Info("The document is not valid, Errors: %v", result.Errors())
		for _, desc := range result.Errors() {
			field := desc.Field()
			validationError := desc.Description()
			errList = append(errList, &models.ConfigValidationError{Field: &field, Error: &validationError})
		}
		return errList, errors.New("Validation failed!")
	}
	return
}

func (cm *CM) ReadFile(name string, data interface{}) (err error) {
	f, err := ioutil.ReadFile(name)
	if err != nil {
		appmgr.Logger.Info("Reading '%s' file failed: %v", name, err)
		return
	}

	err = json.Unmarshal(f, &data)
	if err != nil {
		appmgr.Logger.Info("Unmarshalling '%s' file failed: %v", name, err)
		return
	}

	return
}

func (cm *CM) FetchChart(name string) (err error) {
	tarDir := viper.GetString("xapp.tarDir")
	repo := viper.GetString("helm.repo-name")
	fetchArgs := fmt.Sprintf("--untar --untardir %s %s/%s", tarDir, repo, name)

	_, err = util.HelmExec(strings.Join([]string{"fetch ", fetchArgs}, ""))
	return
}

func (cm *CM) GetMessages(name string) (msgs appmgr.MessageTypes) {
	appmgr.Logger.Info("Fetching tx/rx messages for: %s", name)

	ns := cm.GetNamespace("")
	args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s", ns, cm.GetConfigMapName(name, ns))
	out, err := util.KubectlExec(args)
	if err != nil {
		return
	}

	var p fastjson.Parser
	v, err := p.Parse(string(out))
	if err != nil {
		appmgr.Logger.Info("fastjson.Parser for '%s' failed: %v", name, err)
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

func (cm *CM) GetConfigMapName(xappName, namespace string) string {
	return " configmap-" + namespace + "-" + xappName + "-appconfig"
}

func (cm *CM) GetNamespace(ns string) string {
	if ns != "" {
		return ns
	}

	ns = viper.GetString("xapp.namespace")
	if ns == "" {
		ns = "ricxapp"
	}
	return ns
}
