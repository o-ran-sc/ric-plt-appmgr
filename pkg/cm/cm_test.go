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

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/util"
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
       "rxMessages": ["RIC_X2_LOAD_INFORMATION"]
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

func (cm *MockedConfigMapper) CreateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	return
}

func (cm *MockedConfigMapper) GetConfigMap(m models.XappDescriptor, c *interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) UpdateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error) {
	return
}

func (cm *MockedConfigMapper) DeleteConfigMap(r models.XAppConfig) (c interface{}, err error) {
	return
}

func (cm *MockedConfigMapper) PurgeConfigMap(m models.XappDescriptor) (c interface{}, err error) {
	return
}

func (cm *MockedConfigMapper) RestoreConfigMap(m models.XappDescriptor, c interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) ReadConfigMap(name string, ns string, c *interface{}) (err error) {
	return
}

func (cm *MockedConfigMapper) ApplyConfigMap(r models.XAppConfig, action string) (err error) {
	return
}

func (cm *MockedConfigMapper) FetchChart(name string) (err error) {
	return
}

func (cm *MockedConfigMapper) GetMessages(name string) (msgs appmgr.MessageTypes) {
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

func TestGetMessages(t *testing.T) {
	expectedMsgs := appmgr.MessageTypes{
		TxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		RxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
	}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(kubectlConfigmapOutput), nil
	}

	result := NewCM().GetMessages("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("TestGetMessages failed: expected: %v, got: %v", expectedMsgs, result)
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

func TestApplyConfigMapSuccess(t *testing.T) {
	name := "dummy-xapp"
	m := models.ConfigMetadata{Name: &name, Namespace: "ricxapp"}
	s := ConfigSample{5, "localhost"}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	err := NewCM().ApplyConfigMap(models.XAppConfig{Metadata: &m, Config: s}, "create")
	if err != nil {
		t.Errorf("ApplyConfigMap failed: %v", err)
	}
}

func TestRestoreConfigMapSuccess(t *testing.T) {
	name := "dummy-xapp"
	m := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}
	s := ConfigSample{5, "localhost"}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	err := NewCM().RestoreConfigMap(m, s)
	if err != nil {
		t.Errorf("RestoreConfigMap failed: %v", err)
	}
}

func TestDeleteConfigMapSuccess(t *testing.T) {
	util.HelmExec = func(args string) (out []byte, err error) {
		return []byte("ok"), nil
	}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	validationErrors, err := NewCM().DeleteConfigMap(models.ConfigMetadata{})
	if err != nil {
		t.Errorf("DeleteConfigMap failed: %v -> %v", err, validationErrors)
	}
}

func TestPurgeConfigMapSuccess(t *testing.T) {
	util.HelmExec = func(args string) (out []byte, err error) {
		return []byte("ok"), nil
	}

	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte(`{"logger": {"level": 2}}`), nil
	}

	name := "dummy-xapp"
	validationErrors, err := NewCM().PurgeConfigMap(models.XappDescriptor{XappName: &name})
	if err != nil {
		t.Errorf("PurgeConfigMap failed: %v -> %v", err, validationErrors)
	}
}

func TestCreateConfigMapFails(t *testing.T) {
	name := "dummy-xapp"
	validationErrors, err := NewCM().CreateConfigMap(models.XAppConfig{Metadata: &models.ConfigMetadata{Name: &name}})
	if err == nil {
		t.Errorf("CreateConfigMap failed: %v -> %v", err, validationErrors)
	}
}

func TestUpdateConfigMapFails(t *testing.T) {
	name := "dummy-xapp"
	validationErrors, err := NewCM().UpdateConfigMap(models.XAppConfig{Metadata: &models.ConfigMetadata{Name: &name}})
	if err == nil {
		t.Errorf("CreateConfigMap failed: %v -> %v", err, validationErrors)
	}
}

func TestValidationSuccess(t *testing.T) {
	var d interface{}
	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(`{"local": {"host": ":8080"}, "logger": {"level": 3}}`), &cfg)

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
	err := json.Unmarshal([]byte(`{"local": {"host": ":8080"}, "logger": {"level": "INVALID"}}`), &cfg)

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
