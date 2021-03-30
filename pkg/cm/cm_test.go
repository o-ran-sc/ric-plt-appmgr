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
	"strings"
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/util"
)

const (
	expectedHelmSearchCmd = "search helm-repo"
	expectedHelmFetchCmd  = `fetch --untar --untardir /tmp helm-repo/dummy-xapp`
)

var caughtKubeExecArgs []string
var kubeExecRetOut string
var kubeExecRetErr error
var caughtHelmExecArgs string
var helmExecRetOut string
var helmExecRetErr error

var expectedKubectlGetCmd []string = []string{
	`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-anr-appconfig`,
	`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-appmgr-appconfig`,
	`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-dualco-appconfig`,
	`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-reporter-appconfig`,
	`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-uemgr-appconfig`,
}

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
var kubectlNewConfigmapOutput = `
{
    "name": "ueec",
    "version": "0.7.0",
    "vendor": "Nokia",
    "moId": "SEP",
    "containers": [
        {
            "name": "ueec",
            "image": {
                "registry": "ranco-dev-tools.eastus.cloudapp.azure.com:10001",
                "name": "ueec-xapp",
                "tag": "0.5.3"
            },
            "resources": {
                "limits": {
                    "cpu": "1",
                    "memory": "50Mi"
                },
                "requests": {
                    "cpu": "1",
                    "memory": "100Mi"
                }
            }
        }
    ],
    "livenessProbe": {
        "httpGet": {
            "path": "ric/v1/health/alive",
            "port": 8080
        },
        "initialDelaySeconds": 5,
        "periodSeconds": 15
    },
    "readinessProbe": {
        "httpGet": {
            "path": "ric/v1/health/ready",
            "port": 8080
        },
        "initialDelaySeconds": 5,
        "periodSeconds": 15
    },
    "messaging": {
        "ports": [
            {
                "name": "http",
                "container": "ueec",
                "port": 8080,
                "description": "http service"
            },
            {
                "name": "rmr-route",
                "container": "ueec",
                "port": 4561,
                "description": "rmr route port for ueec"
            },
            {
                "name": "rmr-data",
                "container": "ueec",
                "port": 4560,
                "maxSize": 2072,
                "threadType": 0,
                "lowLatency": false,
                "txMessages": ["RIC_X2_LOAD_INFORMATION"],
				"rxMessages": ["RIC_X2_LOAD_INFORMATION"],
				"policies":   [11, 22, 33],
                "description": "rmr data port for ueec"
            }
        ]
    },
    "controls": {
        "logger": {
            "level": 3
        },
        "subscription": {
            "subscriptionActive": true,
            "functionId": 1,
            "plmnId": "310150",
            "eNBId": "202251",
            "timeout": 5,
            "host": "service-ricplt-submgr-http.ricplt:8088",
            "clientEndpoint": "service-ricxapp-ueec-http.ricxapp:8080"
        }
    },
    "metrics": {
        "url": "/ric/v1/metrics",
        "namespace": "ricxapp"
    },
    "faults": { },
    "measurements": []
}
`
var cfgData = `{
	"active":true,
	"interfaceId": {
		"globalENBId":{
			"plmnId": "1234",
			"eNBId":"55"
		}
	}
}`

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

func TestUploadConfigAllSuccess(t *testing.T) {
	var cfg interface{}
	var expectedResult models.AllXappConfig
	ns := "ricxapp"
	xapps := []string{"anr", "appmgr", "dualco", "reporter", "uemgr"}

	if ret := json.Unmarshal([]byte(cfgData), &cfg); ret != nil {
		t.Errorf("UploadConfigAll Json unmarshal failed: %v", ret)
	}

	for i, _ := range xapps {
		expectedResult = append(expectedResult,
			&models.XAppConfig{
				Config: cfg,
				Metadata: &models.ConfigMetadata{
					Namespace: &ns,
					XappName:  &xapps[i],
				},
			},
		)
	}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	//Fake helm search success
	helmExecRetOut = helmSearchOutput

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = strings.ReplaceAll(cfgData, "\\", "")

	result := NewCM().UploadConfigAll()
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("UploadConfigAll failed: expected: %v, got: %v", expectedResult, result)
	}
	if caughtHelmExecArgs != expectedHelmSearchCmd {
		t.Errorf("UploadConfigAll failed: expected: %v, got: %v", expectedHelmSearchCmd, caughtHelmExecArgs)
	}
	if !reflect.DeepEqual(caughtKubeExecArgs, expectedKubectlGetCmd) {
		t.Errorf("UploadConfigAll failed: expected: %v, got: %v", expectedKubectlGetCmd, caughtKubeExecArgs)
	}
}

func TestUploadConfigAllReturnsEmptyMapIfAllConfigMapReadsFail(t *testing.T) {
	var expectedResult models.AllXappConfig

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	//Fake helm search success
	helmExecRetOut = helmSearchOutput

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' failure
	kubeExecRetErr = errors.New("some error")

	result := NewCM().UploadConfigAll()
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("UploadConfigAll failed: expected: %v, got: %v", expectedResult, result)
	}
}

func TestUploadConfigElementSuccess(t *testing.T) {
	var cfg interface{}
	var expectedResult models.AllXappConfig
	ns := "ricxapp"
	xapps := []string{"anr", "appmgr", "dualco", "reporter", "uemgr"}

	if ret := json.Unmarshal([]byte(cfgData), &cfg); ret != nil {
		t.Errorf("UploadConfigElement Json unmarshal failed: %v", ret)
	}

	for i, _ := range xapps {
		expectedResult = append(expectedResult,
			&models.XAppConfig{
				Config: cfg.(map[string]interface{})["active"],
				Metadata: &models.ConfigMetadata{
					Namespace: &ns,
					XappName:  &xapps[i],
				},
			},
		)
	}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	//Fake helm search success
	helmExecRetOut = helmSearchOutput

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = strings.ReplaceAll(cfgData, "\\", "")

	result := NewCM().UploadConfigElement("active")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("UploadConfigElement failed: expected: %v, got: %v", expectedResult, result)
	}
	if caughtHelmExecArgs != expectedHelmSearchCmd {
		t.Errorf("UploadConfigElement failed: expected: %v, got: %v", expectedHelmSearchCmd, caughtHelmExecArgs)
	}
	if !reflect.DeepEqual(caughtKubeExecArgs, expectedKubectlGetCmd) {
		t.Errorf("UploadConfigElement failed: expected: %v, got: %v", expectedKubectlGetCmd, caughtKubeExecArgs)
	}
}

func TestUploadConfigElementReturnsEmptyMapIfElementMissingFromConfigMap(t *testing.T) {
	var expectedResult models.AllXappConfig

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	//Fake helm search success
	helmExecRetOut = helmSearchOutput

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = strings.ReplaceAll(cfgData, "\\", "")

	//Try to upload non-existing configuration element
	result := NewCM().UploadConfigElement("some-not-existing-element")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("UploadConfigElement failed: expected: %v, got: %v", expectedResult, result)
	}
}

func TestGetRtmDataSuccess(t *testing.T) {
	expectedKubeCmd := []string{
		`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-dummy-xapp-appconfig`,
	}
	expectedMsgs := appmgr.RtmData{
		TxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		RxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		Policies:   []int64{11, 22, 33},
	}

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = kubectlConfigmapOutput

	result := NewCM().GetRtmData("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("GetRtmData failed: expected: %v, got: %v", expectedMsgs, result)
	}
	if !reflect.DeepEqual(caughtKubeExecArgs, expectedKubeCmd) {
		t.Errorf("GetRtmData failed: expected: '%v', got: '%v'", expectedKubeCmd, caughtKubeExecArgs)
	}
}

func TestGetRtmDataNewSuccess(t *testing.T) {
	expectedKubeCmd := []string{
		`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-dummy-xapp-appconfig`,
	}
	expectedMsgs := appmgr.RtmData{
		TxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		RxMessages: []string{"RIC_X2_LOAD_INFORMATION"},
		Policies:   []int64{11, 22, 33},
	}

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = kubectlNewConfigmapOutput

	result := NewCM().GetRtmData("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("GetRtmData failed: expected: %v, got: %v", expectedMsgs, result)
	}
	if !reflect.DeepEqual(caughtKubeExecArgs, expectedKubeCmd) {
		t.Errorf("GetRtmData failed: expected: '%v', got: '%v'", expectedKubeCmd, caughtKubeExecArgs)
	}
}

func TestGetRtmDataReturnsNoDataIfConfigmapGetFails(t *testing.T) {
	var expectedMsgs appmgr.RtmData

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' failure
	kubeExecRetErr = errors.New("some error")

	result := NewCM().GetRtmData("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("GetRtmData failed: expected: %v, got: %v", expectedMsgs, result)
	}
}

func TestGetRtmDataReturnsNoDataIfJsonParseFails(t *testing.T) {
	var expectedMsgs appmgr.RtmData

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' to return nothing what will cause JSON parse failure

	result := NewCM().GetRtmData("dummy-xapp")
	if !reflect.DeepEqual(result, expectedMsgs) {
		t.Errorf("GetRtmData failed: expected: %v, got: %v", expectedMsgs, result)
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
	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec

	if NewCM().FetchChart("dummy-xapp") != nil {
		t.Errorf("TestFetchChart failed!")
	}
}

func TestGetNamespaceSuccess(t *testing.T) {
	if ns := NewCM().GetNamespace("my-ns"); ns != "my-ns" {
		t.Errorf("GetNamespace failed: expected: my-ns, got: %s", ns)
	}
}

func TestGetNamespaceReturnsConfiguredNamespaceName(t *testing.T) {
	if ns := NewCM().GetNamespace(""); ns != "ricxapp" {
		t.Errorf("GetNamespace failed: expected: ricxapp, got: %s", ns)
	}
}

func TestGetNamesFromHelmRepoSuccess(t *testing.T) {
	expectedResult := []string{"anr", "appmgr", "dualco", "reporter", "uemgr"}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	//Fake helm search success
	helmExecRetOut = helmSearchOutput

	names := NewCM().GetNamesFromHelmRepo()
	if !reflect.DeepEqual(names, expectedResult) {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
	if caughtHelmExecArgs != expectedHelmSearchCmd {
		t.Errorf("GetNamesFromHelmRepo failed: expected: %v, got: %v", expectedHelmSearchCmd, caughtHelmExecArgs)
	}
}

func TestGetNamesFromHelmRepoFailure(t *testing.T) {
	expectedResult := []string{}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	helmExecRetOut = helmSearchOutput
	helmExecRetErr = errors.New("Command failed!")

	EnvHelmVersion = "3"
	names := NewCM().GetNamesFromHelmRepo()
	if names != nil {
		t.Errorf("GetNamesFromHelmRepo failed: expected %v, got %v", expectedResult, names)
	}
}

func TestBuildConfigMapSuccess(t *testing.T) {
	expectedKubeCmd := []string{
		`get configmap -o jsonpath='{.data.config-file\.json}' -n ricxapp  configmap-ricxapp-dummy-xapp-appconfig`,
	}
	name := "dummy-xapp"
	namespace := "ricxapp"
	m := models.ConfigMetadata{XappName: &name, Namespace: &namespace}
	s := `{"Metadata": {"XappName": "ueec", "Namespace": "ricxapp"}, ` +
		`"Config": {"active": true, "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}}`

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' success
	kubeExecRetOut = `{"logger": {"level": 2}}`

	cmString, err := NewCM().BuildConfigMap(models.XAppConfig{Metadata: &m, Config: s})
	if err != nil {
		t.Errorf("BuildConfigMap failed: %v -> %v", err, cmString)
	}
	if !reflect.DeepEqual(caughtKubeExecArgs, expectedKubeCmd) {
		t.Errorf("BuildConfigMap failed: expected: %v, got: %v", expectedKubeCmd, caughtKubeExecArgs)
	}
}

func TestBuildConfigMapReturnErrorIfJsonMarshalFails(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	m := models.ConfigMetadata{XappName: &name, Namespace: &namespace}
	//Give channel as a configuration input, this will fail JSON marshal
	cmString, err := NewCM().BuildConfigMap(models.XAppConfig{Metadata: &m, Config: make(chan int)})
	if err == nil {
		t.Errorf("BuildConfigMap failed: %v -> %v", err, cmString)
	}
}

func TestBuildConfigMapReturnErrorIfKubectlGetConfigmapFails(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	m := models.ConfigMetadata{XappName: &name, Namespace: &namespace}
	s := `{"Metadata": {"XappName": "ueec", "Namespace": "ricxapp"}, ` +
		`"Config": {"active": true, "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}}`

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl get configmap' failure
	kubeExecRetErr = errors.New("some error")

	cmString, err := NewCM().BuildConfigMap(models.XAppConfig{Metadata: &m, Config: s})
	if err == nil {
		t.Errorf("BuildConfigMap failed: %v -> %v", err, cmString)
	} else if err.Error() != "some error" {
		t.Errorf("BuildConfigMap failed: expected: 'some error', got: '%s'", err.Error())
	}
}

func TestBuildConfigMapReturnErrorIfJsonParserFails(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	m := models.ConfigMetadata{XappName: &name, Namespace: &namespace}
	s := `{"Metadata": {"XappName": "ueec", "Namespace": "ricxapp"}, ` +
		`"Config": {"active": true, "interfaceId":{"globalENBId": {"eNBId": 77, "plmnId": "6666"}}}}`

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Return empty json that causes JSON parser to fail
	kubeExecRetOut = ``

	cmString, err := NewCM().BuildConfigMap(models.XAppConfig{Metadata: &m, Config: s})
	if err == nil {
		t.Errorf("BuildConfigMap failed: %v -> %v", err, cmString)
	}
}

func TestGenerateJSONFileSuccess(t *testing.T) {
	err := NewCM().GenerateJSONFile("{}")
	if err != nil {
		t.Errorf("GenerateJSONFile failed: %v", err)
	}
}

func TestReplaceConfigMapSuccess(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"

	defer func() { resetKubeExecMock() }()
	kubeExec = mockedKubeExec
	//Fake 'kubectl create configmap' success
	kubeExecRetOut = ""

	err := NewCM().ReplaceConfigMap(name, namespace)
	if err != nil {
		t.Errorf("ReplaceConfigMap failed: %v", err)
	}
}

func TestUpdateConfigMapReturnsErrorIfSchemaFileIsMissing(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	config := models.XAppConfig{Metadata: &models.ConfigMetadata{XappName: &name, Namespace: &namespace}}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	helmExecRetOut = `{}`

	//Will fail at schema reading, because schema file is mission
	validationErrors, err := NewCM().UpdateConfigMap(config)
	if err == nil {
		t.Errorf("UpdateConfigMap failed: %v -> %v", err, validationErrors)
	}
	if caughtHelmExecArgs != expectedHelmFetchCmd {
		t.Errorf("UpdateConfigMap failed: expected: %v, got: %v", expectedHelmFetchCmd, caughtHelmExecArgs)
	}
}

func TestUpdateConfigMapReturnsErrorIfHelmFetchChartFails(t *testing.T) {
	name := "dummy-xapp"
	namespace := "ricxapp"
	config := models.XAppConfig{Metadata: &models.ConfigMetadata{XappName: &name, Namespace: &namespace}}

	defer func() { resetHelmExecMock() }()
	helmExec = mockedHelmExec
	helmExecRetErr = errors.New("some error")

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
		t.Errorf("doValidate should fail but didn't: %v -> %v", err, feedback)
	}
	appmgr.Logger.Debug("Feedbacks: %v", feedback)
}

func TestReadFileReturnsErrorIfFileReadFails(t *testing.T) {
	var d interface{}

	if err := NewCM().ReadFile("not/existing/test/schema.json", &d); err == nil {
		t.Errorf("ReadFile should fail but it didn't")
	}
}

func TestReadFileReturnsErrorIfJsonUnmarshalFails(t *testing.T) {
	var d interface{}

	if err := NewCM().ReadFile("../../test/faulty_schema.json", &d); err == nil {
		t.Errorf("ReadFile should fail but it didn't")
	}
}

func mockedKubeExec(args string) (out []byte, err error) {
	caughtKubeExecArgs = append(caughtKubeExecArgs, args)
	return []byte(kubeExecRetOut), kubeExecRetErr
}

func resetKubeExecMock() {
	kubeExec = util.KubectlExec
	caughtKubeExecArgs = nil
	kubeExecRetOut = ""
	kubeExecRetErr = nil
}

func mockedHelmExec(args string) (out []byte, err error) {
	caughtHelmExecArgs = args
	return []byte(helmExecRetOut), helmExecRetErr
}

func resetHelmExecMock() {
	helmExec = util.HelmExec
	caughtHelmExecArgs = ""
	helmExecRetOut = ""
	helmExecRetErr = nil
}
