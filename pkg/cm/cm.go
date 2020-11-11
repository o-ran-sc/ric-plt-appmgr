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
        "strconv"
        "strings"

        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/util"
)

var kubeExec = util.KubectlExec
var helmExec = util.HelmExec

type CM struct{}

const HELM_VERSION_3 = "3"
const HELM_VERSION_2 = "2"
var EnvHelmVersion string = ""


func NewCM() *CM {
        return &CM{}
}

func (cm *CM) UploadConfigAll() (configList models.AllXappConfig) {
        return cm.UploadConfigElement("")
}

func (cm *CM) UploadConfigElement(Element string) (configList models.AllXappConfig) {
        namespace := cm.GetNamespace("")
        for _, name := range cm.GetNamesFromHelmRepo() {
                var activeConfig interface{}
                xAppName := name
                if err := cm.GetConfigmap(xAppName, namespace, &activeConfig); err != nil {
                        appmgr.Logger.Info("No active configMap found for '%s', ignoring ...", xAppName)
                        continue
                }

                if Element != "" {
                        m := activeConfig.(map[string]interface{})
                        if m[Element] == nil {
                                appmgr.Logger.Info("xApp '%s' doesn't have requested element '%s' in config", name, Element)
                                continue
                        }
                        activeConfig = m[Element]
                }

                c := models.XAppConfig{
                        Metadata: &models.ConfigMetadata{XappName: &xAppName, Namespace: &namespace},
                        Config:   activeConfig,
                }
                configList = append(configList, &c)
        }
        return
}

func (cm *CM) GetConfigmap(name, namespace string, c *interface{}) (err error) {
        cmJson, err := cm.ReadConfigmap(name, namespace)
        if err != nil {
                return err
        }

        return json.Unmarshal([]byte(cmJson), &c)
}

func (cm *CM) ReadSchema(name string, desc *interface{}) (err error) {
        if err = cm.FetchChart(name); err != nil {
                return
        }

        tarDir := viper.GetString("xapp.tarDir")
        err = cm.ReadFile(path.Join(tarDir, name, viper.GetString("xapp.schema")), desc)
        if err != nil {
                return
        }

        if err = os.RemoveAll(path.Join(tarDir, name)); err != nil {
                appmgr.Logger.Info("RemoveAll failed: %v", err)
        }

        return
}

func (cm *CM) UpdateConfigMap(r models.XAppConfig) (models.ConfigValidationErrors, error) {
        fmt.Printf("Configmap update: xappName=%s namespace=%s config: %v\n", *r.Metadata.XappName, *r.Metadata.Namespace, r.Config)
        if validationErrors, err := cm.Validate(r); err != nil {
                return validationErrors, err
        }

        cmContent, err := cm.BuildConfigMap(r)
        if err != nil {
                return nil, err
        }

        if err := cm.GenerateJSONFile(cmContent); err != nil {
                return nil, err
        }
        err = cm.ReplaceConfigMap(*r.Metadata.XappName, *r.Metadata.Namespace)

        return nil, err
}

func (cm *CM) BuildConfigMap(r models.XAppConfig) (string, error) {
        configJson, err := json.Marshal(r.Config)
        if err != nil {
                appmgr.Logger.Info("Config marshalling failed: %v", err)
                return "", err
        }

        cmContent, err := cm.ReadConfigmap(*r.Metadata.XappName, *r.Metadata.Namespace)
        if err != nil {
                return "", err
        }

        v, err := cm.ParseJson(cmContent)
        if err == nil {
                v.Set("controls", fastjson.MustParse(string(configJson)))
                fmt.Println(v.String())
                return v.String(), nil
        }

        return "", err
}

func (cm *CM) ParseJson(dsContent string) (*fastjson.Value, error) {
        var p fastjson.Parser
        v, err := p.Parse(dsContent)
        if err != nil {
                appmgr.Logger.Info("fastjson.Parser failed: %v", err)
        }
        return v, err
}

func (cm *CM) GenerateJSONFile(jsonString string) error {
        cmJson, err := json.RawMessage(jsonString).MarshalJSON()
        if err != nil {
                appmgr.Logger.Error("Config marshalling failed: %v", err)
                return err
        }

        err = ioutil.WriteFile(viper.GetString("xapp.tmpConfig"), cmJson, 0644)
        if err != nil {
                appmgr.Logger.Error("WriteFile failed: %v", err)
                return err
        }

        return nil
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

func (cm *CM) ReadConfigmap(name string, ns string) (string, error) {
        args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s", ns, cm.GetConfigMapName(name, ns))
        out, err := kubeExec(args)
        return string(out), err
}

func (cm *CM) ReplaceConfigMap(name, ns string) error {
        cmd := " create configmap -n %s %s --from-file=%s -o json --dry-run | kubectl replace -f -"
        args := fmt.Sprintf(cmd, ns, cm.GetConfigMapName(name, ns), viper.GetString("xapp.tmpConfig"))
        _, err := kubeExec(args)
        return err
}

func (cm *CM) FetchChart(name string) (err error) {
        tarDir := viper.GetString("xapp.tarDir")
        repo := viper.GetString("helm.repo-name")
        fetchArgs := fmt.Sprintf("--untar --untardir %s %s/%s", tarDir, repo, name)

        _, err = helmExec(strings.Join([]string{"fetch ", fetchArgs}, ""))
        return
}

func (cm *CM) GetRtmData(name string) (msgs appmgr.RtmData) {
        appmgr.Logger.Info("Fetching RT data for xApp=%s", name)

        ns := cm.GetNamespace("")
        args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s", ns, cm.GetConfigMapName(name, ns))
        out, err := kubeExec(args)
        if err != nil {
                return
        }

        var p fastjson.Parser
        v, err := p.Parse(string(out))
        if err != nil {
                appmgr.Logger.Info("fastjson.Parser for '%s' failed: %v", name, err)
                return
        }

        if v.Exists("rmr") {
                for _, m := range v.GetArray("rmr", "txMessages") {
                        msgs.TxMessages = append(msgs.TxMessages, strings.Trim(m.String(), `"`))
                }

                for _, m := range v.GetArray("rmr", "rxMessages") {
                        msgs.RxMessages = append(msgs.RxMessages, strings.Trim(m.String(), `"`))
                }

                for _, m := range v.GetArray("rmr", "policies") {
                        if val, err := strconv.Atoi(strings.Trim(m.String(), `"`)); err == nil {
                                msgs.Policies = append(msgs.Policies, int64(val))
                        }
                }
        } else {
                for _, p := range v.GetArray("messaging", "ports") {
                        appmgr.Logger.Info("txMessages=%v, rxMessages=%v", p.GetArray("txMessages"), p.GetArray("rxMessages"))
                        for _, m := range p.GetArray("txMessages") {
                                msgs.TxMessages = append(msgs.TxMessages, strings.Trim(m.String(), `"`))
                        }

                        for _, m := range p.GetArray("rxMessages") {
                                msgs.RxMessages = append(msgs.RxMessages, strings.Trim(m.String(), `"`))
                        }

                        for _, m := range p.GetArray("policies") {
                                if val, err := strconv.Atoi(strings.Trim(m.String(), `"`)); err == nil {
                                        msgs.Policies = append(msgs.Policies, int64(val))
                                }
                        }
                }
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

func (cm *CM) GetNamesFromHelmRepo() (names []string) {
        rname := viper.GetString("helm.repo-name")

        var cmdArgs string = ""
        if EnvHelmVersion == HELM_VERSION_3 {
                cmdArgs = strings.Join([]string{"search repo ", rname}, "")
        }else {
                 cmdArgs = strings.Join([]string{"search ", rname}, "")
        }

        out, err := helmExec(cmdArgs)
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
        var desc interface{}
        err = cm.ReadSchema(*req.Metadata.XappName, &desc)
        if err != nil {
                appmgr.Logger.Info("No schema file found for '%s', aborting ...", *req.Metadata.XappName)
                return
        }
        return cm.doValidate(desc, req.Config)
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
        appmgr.Logger.Info("Config validation successful!")

        return
}
