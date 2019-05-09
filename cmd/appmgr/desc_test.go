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
	"testing"
	"reflect"
	"errors"
	"encoding/json"
	"log"
)

var helmSearchOutput = `
helm-repo/anr           0.0.1           1.0             Helm Chart for Nokia ANR (Automatic Neighbour Relation) xAPP
helm-repo/appmgr        0.0.2           1.0             Helm Chart for xAppManager
helm-repo/dualco        0.0.1           1.0             Helm Chart for Nokia dualco xAPP
helm-repo/reporter      0.0.1           1.0             Helm Chart for Reporting xAPP
helm-repo/uemgr         0.0.1           1.0             Helm Chart for Nokia uemgr xAPP
`
type ConfigSample struct {
	Level  	int
	Host 	string
}

type MockedConfigMapper struct {
}

func (cm *MockedConfigMapper) UploadConfig() (cfg []XAppConfig) {
	return
}

func (cm *MockedConfigMapper) CreateConfigMap(r XAppConfig) (errList []CMError, err error){
	return
}

func (cm *MockedConfigMapper) UpdateConfigMap(r XAppConfig) (errList []CMError, err error){
	return
}

func (cm *MockedConfigMapper) DeleteConfigMap(r XAppConfig) (c interface{}, err error){
	return
}

func (cm *MockedConfigMapper) PurgeConfigMap(m XappDeploy) (c interface{}, err error){
	return
}

func (cm *MockedConfigMapper) RestoreConfigMap(m XappDeploy, c interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) ReadConfigMap(name string, ns string, c *interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) ApplyConfigMap(r XAppConfig, action string) (err error) {
	return
}

func (cm *MockedConfigMapper) FetchChart(name string) (err error) {
	return
}

func (cm *MockedConfigMapper) GetMessages(name string) (msgs MessageTypes) {
	return
}

// Test cases
func TestGetMessages(t *testing.T) {
	cm := ConfigMap{}
	expectedMsgs := MessageTypes{}

	if !reflect.DeepEqual(cm.GetMessages("dummy-xapp"), expectedMsgs) {
		t.Errorf("TestGetMessages failed!")
	}
}

func TestFetchChartFails(t *testing.T) {
	cm := ConfigMap{}

	if cm.FetchChart("dummy-xapp") == nil {
		t.Errorf("TestFetchChart failed!")
	}
}

func TestFetchChartSuccess(t *testing.T) {
	cm := ConfigMap{}

	HelmExec = func(args string) (out []byte, err error) {
		return
	}

	if cm.FetchChart("dummy-xapp") != nil {
		t.Errorf("TestFetchChart failed!")
	}
}

func TestGetNamesFromHelmRepoSuccess(t *testing.T) {
	cm := ConfigMap{}
	expectedResult := []string{"anr", "appmgr", "dualco", "reporter", "uemgr"}
	HelmExec = func(args string) (out []byte, err error) {
		return []byte(helmSearchOutput), nil
	}

	names := cm.GetNamesFromHelmRepo()
	if !reflect.DeepEqual(names, expectedResult) {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
}

func TestGetNamesFromHelmRepoFailure(t *testing.T) {
	cm := ConfigMap{}
	expectedResult := []string{}
	HelmExec = func(args string) (out []byte, err error) {
		return []byte(helmSearchOutput), errors.New("Command failed!")
	}

	names := cm.GetNamesFromHelmRepo()
	if names != nil {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
}

func TestApplyConfigMapSuccess(t *testing.T) {
	cm := ConfigMap{}
	m := ConfigMetadata{Name: "dummy-xapp", Namespace: "ricxapp"}
	s := ConfigSample{5, "localhost"}

	KubectlExec = func(args string) (out []byte, err error) {
		log.Println("TestApplyConfigMapSuccess: ", args)
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	err := cm.ApplyConfigMap(XAppConfig{Metadata: m, Configuration: s}, "create")
	if err != nil {
		t.Errorf("ApplyConfigMap failed: %v", err)
	}
}

func TestRestoreConfigMapSuccess(t *testing.T) {
	cm := ConfigMap{}
	m := XappDeploy{Name: "dummy-xapp", Namespace: "ricxapp"}
	s := ConfigSample{5, "localhost"}

	KubectlExec = func(args string) (out []byte, err error) {
		log.Println("TestRestoreConfigMapSuccess: ", args)
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	err := cm.RestoreConfigMap(m, s)
	if err != nil {
		t.Errorf("RestoreConfigMap failed: %v", err)
	}
}

func TestDeleteConfigMapSuccess(t *testing.T) {
	cm := ConfigMap{}

	HelmExec = func(args string) (out []byte, err error) {
		return []byte("ok"), nil
	}

	KubectlExec = func(args string) (out []byte, err error) {
		log.Println("TestDeleteConfigMapSuccess: ", args)
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	c, err := cm.DeleteConfigMap(XAppConfig{})
	if err != nil {
		t.Errorf("DeleteConfigMap failed: %v -> %v", err, c)
	}
}

func TestPurgeConfigMapSuccess(t *testing.T) {
	cm := ConfigMap{}

	HelmExec = func(args string) (out []byte, err error) {
		return []byte("ok"), nil
	}

	KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	c, err := cm.PurgeConfigMap(XappDeploy{})
	if err != nil {
		t.Errorf("PurgeConfigMap failed: %v -> %v", err, c)
	}
}

func TestCreateConfigMapFails(t *testing.T) {
	cm := ConfigMap{}

	c, err := cm.CreateConfigMap(XAppConfig{})
	if err == nil {
		t.Errorf("CreateConfigMap failed: %v -> %v", err, c)
	}
}

func TestUpdateConfigMapFails(t *testing.T) {
	cm := ConfigMap{}

	c, err := cm.UpdateConfigMap(XAppConfig{})
	if err == nil {
		t.Errorf("CreateConfigMap failed: %v -> %v", err, c)
	}
}

func TestValidationSuccess(t *testing.T) {
	cm := ConfigMap{}
	var d interface{}
	var cfg map[string]interface{}

	err := json.Unmarshal([]byte(`{"local": {"host": ":8080"}, "logger": {"level": 3}}`), &cfg)

	err = cm.ReadFile("./test/schema.json", &d)
	if err != nil {
		t.Errorf("ReadFile failed: %v -> %v", err, d)
	}

	feedback, err := cm.doValidate(d, cfg)
	if err != nil {
		t.Errorf("doValidate failed: %v -> %v", err, feedback)
	}
}

func TestValidationFails(t *testing.T) {
	cm := ConfigMap{}
	var d interface{}
	var cfg map[string]interface{}

	err := json.Unmarshal([]byte(`{"local": {"host": ":8080"}, "logger": {"level": "INVALID"}}`), &cfg)

	err = cm.ReadFile("./test/schema.json", &d)
	if err != nil {
		t.Errorf("ConfigMetadata failed: %v -> %v", err, d)
	}

	feedback, err := cm.doValidate(d, cfg)
	if err == nil {
		t.Errorf("doValidate should faile but didn't: %v -> %v", err, feedback)
	}

	log.Println("Feedbacks: ", feedback)
}