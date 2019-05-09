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

func UploadConfig() (cfg []XAppConfig) {
	for _, name := range GetNamesFromHelmRepo() {
		if name == "appmgr" {
			continue
		}

		c := XAppConfig{
			Metadata: ConfigMetadata{Name: name, Namespace: "ricxapp", ConfigName: name + "-appconfig"},
		}

		err := ReadSchema(name, &c)
		if err != nil {
			continue
		}

		err = ReadConfigMap(name, "ricxapp", &c.Configuration)
		if err != nil {
			log.Println("No active configMap found, using default!")
		}

		cfg = append(cfg, c)
	}
	return
}

func ReadSchema(name string, c *XAppConfig) (err error) {
	if err = FetchChart(name); err != nil {
		return
	}

	tarDir := viper.GetString("xapp.tarDir")
	err = ReadFile(path.Join(tarDir, name, viper.GetString("xapp.schema")), &c.Descriptor)
	if err != nil {
		return
	}

	err = ReadFile(path.Join(tarDir, name, viper.GetString("xapp.config")), &c.Configuration)
	if err != nil {
		return
	}

	if err = os.RemoveAll(path.Join(tarDir, name)); err != nil {
		log.Println("RemoveAll failed", err)
	}

	return
}

func ReadConfigMap(name string, ns string, c *interface{}) (err error) {
	args := fmt.Sprintf("get configmap -o jsonpath='{.data.config-file\\.json}' -n %s %s-appconfig", ns, name)
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

func ApplyConfigMap(r XAppConfig) (err error) {
	cm := ConfigMap{
		Kind:       "ConfigMap",
		ApiVersion: "v1",
		Metadata:   CMMetadata{Name: r.Metadata.Name, Namespace: r.Metadata.Namespace},
		Data:       r.Configuration,
	}

	cmJson, err := json.Marshal(cm)
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

	cmd := " create configmap -n %s %s --from-file=%s -o json --dry-run | kubectl apply -f -"
	args := fmt.Sprintf(cmd, r.Metadata.Namespace, r.Metadata.ConfigName, cmFile)
	_, err = KubectlExec(args)
	if err != nil {
		return
	}
	log.Println("Configmap changes created!")

	return
}

func CreateConfigMap(r XAppConfig) (err error) {
	if err = Validate(r); err != nil {
		return
	}
	return ApplyConfigMap(r)
}

func DeleteConfigMap(r XAppConfig) (cm interface{}, err error) {
	err = ReadConfigMap(r.Metadata.Name, r.Metadata.Namespace, &cm)
	if err == nil {
		args := fmt.Sprintf(" delete configmap --namespace=%s %s", r.Metadata.Namespace, r.Metadata.ConfigName)
		_, err = KubectlExec(args)
	}
	return
}

func PurgeConfigMap(m ConfigMetadata) (cm interface{}, err error) {
	if m.ConfigName == "" {
		m.ConfigName = m.Name + "-appconfig"
	}
	return DeleteConfigMap(XAppConfig{Metadata: m})
}

func RestoreConfigMap(m ConfigMetadata, cm interface{}) (err error) {
	if m.ConfigName == "" {
		m.ConfigName = m.Name + "-appconfig"
	}
	time.Sleep(time.Duration(10 * time.Second))

	return ApplyConfigMap(XAppConfig{Metadata: m, Configuration: cm})
}

func GetNamesFromHelmRepo() (names []string) {
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

func Validate(req XAppConfig) (err error) {
	c := XAppConfig{}
	err = ReadSchema(req.Metadata.Name, &c)
	if err != nil {
		log.Printf("No schema file found for '%s', aborting ...", req.Metadata.Name)
		return err
	}

	schemaLoader := gojsonschema.NewGoLoader(c.Descriptor)
	documentLoader := gojsonschema.NewGoLoader(req.Configuration)

	log.Println("Starting validation ...")
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Println("Validation failed: ", err)
		return
	}

	log.Println("validation done ...", err, result.Valid())
	if result.Valid() == false {
		log.Println("The document is not valid, Errors: ", result.Errors())
		s := make([]string, 3)
		for i, desc := range result.Errors() {
			s = append(s, fmt.Sprintf(" (%d): %s.\n", i, desc.String()))
		}
		return errors.New(strings.Join(s, " "))
	}
	return
}

func ReadFile(name string, data interface{}) (err error) {
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

func FetchChart(name string) (err error) {
	tarDir := viper.GetString("xapp.tarDir")
	repo := viper.GetString("helm.repo-name")
	fetchArgs := fmt.Sprintf("--untar --untardir %s %s/%s", tarDir, repo, name)

	_, err = HelmExec(strings.Join([]string{"fetch ", fetchArgs}, ""))
	return
}

func GetMessages(name string) (msgs MessageTypes, err error) {
	log.Println("Fetching tx/rx messages for: ", name)
	return
}
