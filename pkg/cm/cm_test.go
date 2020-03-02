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
	"os"
	"reflect"
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/util"
)

var helmSearchOutput = `
helm-repo/anr           0.0.1           1.0             Helm Chart for Nokia ANR (Automatic Neighbour Relation) xAPP
helm-repo/appmgr        0.0.2           1.0             Helm Chart for xAppManager
helm-repo/dualco        0.0.1           1.0             Helm Chart for Nokia dualco xAPP
helm-repo/reporter      0.0.1           1.0             Helm Chart for Reporting xAPP
helm-repo/uemgr         0.0.1           1.0             Helm Chart for Nokia uemgr xAPP
`

var kubectlConfigmapOutput = `
{
    "local": {
        "host": ":8080"
    },
    "logger": {
        "level": 3
    },
    "rmr": {
       "protPort": "tcp:4560",
       "maxSize": 2072,
       "numWorkers": 1,
       "txMessages": ["RIC_X2_LOAD_INFORMATION"],
       "rxMessages": ["RIC_X2_LOAD_INFORMATION"],
	   "policies":   [11, 22, 33]
    },
    "db": {
        "namespace": "ricxapp",
        "host": "dbaas",
        "port": 6379
    }
}
`

type ConfigSample struct {
	Level int
	Host  string
}

type MockedConfigMapper struct {
}

func (cm *MockedConfigMapper) ReadSchema(name string, c *models.XAppConfig) (err error) {
	return
}

func (cm *MockedConfigMapper) UploadConfig() (cfg []models.XAppConfig) {
	return
}

func (cm *MockedConfigMapper) UpdateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	return
}

func (cm *MockedConfigMapper) ReadConfigMap(name string, ns string, c *interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) FetchChart(name string) (err error) {
	return
}

func (cm *MockedConfigMapper) GetRtmData(name string) (msgs appmgr.RtmData) {
	return
}

func (cm *MockedConfigMapper) GetNamespace(ns string) (n string) {
	return
}

func (cm *MockedConfigMapper) GetNamesFromHelmRepo() (names []string) {
	return
}

// Test cases
func TestMain(m *testing.M) {
	appmgr.Init()
	appmgr.Logger.SetLevel(0)

	code := m.Run()
	os.Exit(code)
}

func TestGetRtmData(t *testing.T) {
	expectedMsgs := appmgr.RtmData{
		TxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		RxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		Policies:   []int64{11, 22, 33},
	}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(kubectlConfigmapOutput), nil
	}

	result := NewCM().GetRtmData("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("TestGetRtmData failed: expected: %v, got: %v", expectedMsgs, result)
	}
}

func TestHelmNamespace(t *testing.T) {
	if NewCM().GetNamespace("pltxapp") != "pltxapp" {
		t.Errorf("getNamespace failed!")
	}

	if NewCM().GetNamespace("") != "ricxapp" {
		t.Errorf("getNamespace failed!")
	}
}

func TestFetchChartFails(t *testing.T) {
	if NewCM().FetchChart("dummy-xapp") == nil {
		t.Errorf("TestFetchChart failed!")
	}
}

func TestFetchChartSuccess(t *testing.T) {
	util.HelmExec = func(args string) (out []byte, err error) {
		return
	}

	if NewCM().FetchChart("dummy-xapp") != nil {
		t.Errorf("TestFetchChart failed!")
	}
}

func TestGetNamesFromHelmRepoSuccess(t *testing.T) {
	expectedResult := []string{"anr", "appmgr", "dualco", "reporter", "uemgr"}
	util.HelmExec = func(args string) (out []byte, err error) {
		return []byte(helmSearchOutput), nil
	}

	names := NewCM().GetNamesFromHelmRepo()
	if !reflect.DeepEqual(names, expectedResult) {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
}

func TestGetNamesFromHelmRepoFailure(t *testing.T) {
	expectedResult := []string{}
	util.HelmExec = func(args string) (out []byte, err error) {
		return []byte(helmSearchOutput), errors.New("Command failed!")
	}

	names := NewCM().GetNamesFromHelmRepo()
	if names != nil {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
}

func TestBuildConfigMapSuccess(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	m := models.ConfigMetadata{XappName: &name, Namespace: &namespace}
	s := `{"Metadata": {"XappName": "ueec", "Namespace": "ricxapp"}, "Config": {"active": true, "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}}`

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	cmString, err := NewCM().BuildConfigMap(models.XAppConfig{Metadata: &m, Config: s})
	if err != nil {
		t.Errorf("BuildConfigMap failed: %v -> %v", err, cmString)
	}
}

func TestUpdateConfigMapFails(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	config := models.XAppConfig{Metadata: &models.ConfigMetadata{XappName: &name, Namespace: &namespace}}

	validationErrors, err := NewCM().UpdateConfigMap(config)
	if err == nil {
		t.Errorf("UpdateConfigMap failed: %v -> %v", err, validationErrors)
	}
}

func TestValidationSuccess(t *testing.T) {
	var d interface{}
	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(`{"active": true, "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}`), &cfg)

	err = NewCM().ReadFile("../../test/schema.json", &d)
	if err != nil {
		t.Errorf("ReadFile failed: %v -> %v", err, d)
	}

	feedback, err := NewCM().doValidate(d, cfg)
	if err != nil {
		t.Errorf("doValidate failed: %v -> %v", err, feedback)
	}
}

func TestValidationFails(t *testing.T) {
	var d interface{}
	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(`{"active": "INVALID", "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}`), &cfg)

	err = NewCM().ReadFile("../../test/schema.json", &d)
	if err != nil {
		t.Errorf("ConfigMetadata failed: %v -> %v", err, d)
	}

	feedback, err := NewCM().doValidate(d, cfg)
	if err == nil {
		t.Errorf("doValidate should faile but didn't: %v -> %v", err, feedback)
	}
	appmgr.Logger.Debug("Feedbacks: %v", feedback)
}
