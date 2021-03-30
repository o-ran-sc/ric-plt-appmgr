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

package helm

import (
        "errors"
        "github.com/spf13/viper"
        "os"
        "reflect"
        "strconv"
        "strings"
        "testing"
	"github.com/stretchr/testify/assert"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/cm"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
        "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/util"
)

var caughtKubeExecArgs string
var kubeExecRetOut string
var kubeExecRetErr error
var caughtHelmExecArgs string
var helmExecRetOut string
var helmExecRetErr error
var helmStatusOutput = `
LAST DEPLOYED: Sat Mar  9 06:50:45 2019
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Pod(related)
NAME                        READY  STATUS   RESTARTS  AGE
dummy-xapp-8984fc9fd-bkcbp  1/1    Running  0         55m
dummy-xapp-8984fc9fd-l6xch  1/1    Running  0         55m
dummy-xapp-8984fc9fd-pp4hg  1/1    Running  0         55m

==> v1/Service
NAME                         TYPE       CLUSTER-IP      EXTERNAL-IP  PORT(S)  AGE
dummy-xapp-dummy-xapp-chart  ClusterIP  10.102.184.212  <none>       80/TCP   55m

==> v1beta1/Deployment
NAME        READY  UP-TO-DATE  AVAILABLE  AGE
dummy-xapp  3/3    3           3          55m
`

var helListAllOutput = `Next: ""
Releases:
- AppVersion: "1.0"
  Chart: dummy-xapp-chart-0.1.0
  Name: dummy-xapp
  Namespace: default
  Revision: 1
  Status: DEPLOYED
  Updated: Mon Mar 11 06:55:05 2019
- AppVersion: "2.0"
  Chart: dummy-xapp-chart-0.1.0
  Name: dummy-xapp2
  Namespace: default
  Revision: 1
  Status: DEPLOYED
  Updated: Mon Mar 11 06:55:05 2019
- AppVersion: "1.0"
  Chart: appmgr-0.0.1
  Name: appmgr
  Namespace: default
  Revision: 1
  Status: DEPLOYED
  Updated: Sun Mar 24 07:17:00 2019
  `

var helListOutput = `Next: ""
Releases:
- AppVersion: "1.0"
  Chart: dummy-xapp-chart-0.1.0
  Name: dummy-xapp
  Namespace: default
  Revision: 1
  Status: DEPLOYED
  Updated: Mon Mar 11 06:55:05 2019
  `

var kubeServiceOutput = `{
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
        "creationTimestamp": "2020-03-31T12:27:12Z",
        "labels": {
            "app": "ricxapp-dummy-xapp",
            "chart": "dummy-xapp-0.0.4",
            "heritage": "Tiller",
            "release": "dummy-xapp"
        },
        "name": "service-ricxapp-dummy-xapp-rmr",
        "namespace": "ricxapp",
        "resourceVersion": "4423380",
        "selfLink": "/api/v1/namespaces/ricxapp/services/service-ricxapp-dummy-xapp-rmr",
        "uid": "2254b77d-7dd6-43e0-beff-3e2a7b24c89a"
    },
    "spec": {
        "clusterIP": "10.98.239.107",
        "ports": [
            {
                "name": "rmrdata",
                "port": 4560,
                "protocol": "TCP",
                "targetPort": "rmrdata"
            },
            {
                "name": "rmrroute",
                "port": 4561,
                "protocol": "TCP",
                "targetPort": "rmrroute"
            }
        ],
        "selector": {
            "app": "ricxapp-dummy-xapp",
            "release": "dummy-xapp"
        },
        "sessionAffinity": "None",
        "type": "ClusterIP"
    },
    "status": {
        "loadBalancer": {}
    }
}`


// Test cases
func TestMain(m *testing.M) {
        appmgr.Init()
        appmgr.Logger.SetLevel(0)
	
        code := m.Run()
        os.Exit(code)
}

func TestInit(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        var expectedHelmCommand string = ""
        helmExec = mockedHelmExec

        NewHelm().Init()
        if cm.EnvHelmVersion == cm.HELM_VERSION_2{
                expectedHelmCommand = "init -c --skip-refresh"
                if caughtHelmExecArgs != expectedHelmCommand {
                        t.Errorf("Init failed: expected %v, got %v", expectedHelmCommand, caughtHelmExecArgs)
                }
        }
}

func TestAddRepoSuccess(t *testing.T) {
        defer func() {
                resetHelmExecMock()
                removeTestUsernameFile()
                removeTestPasswordFile()
        }()
        helmExec = mockedHelmExec

        if err := writeTestUsernameFile(); err != nil {
                t.Errorf("AddRepo username file create failed: %s", err)
                return
        }
        if err := writeTestPasswordFile(); err != nil {
                t.Errorf("AddRepo password file create failed: %s", err)
                return
        }

        if _, err := NewHelm().AddRepo(); err != nil {
                t.Errorf("AddRepo failed: %v", err)
        }

        if !strings.Contains(caughtHelmExecArgs, "repo add") {
                t.Errorf("AddRepo failed: expected %v, got %v", "repo add", caughtHelmExecArgs)
        }
	NewHelm().initDone = true
	NewHelm().Initialize()
}

func TestFuncsWithHelmv3(t *testing.T){
	var err error
	name := "dymmy-xapp"

	if err = os.Setenv("HELMVERSION", "3"); err != nil { 
	        t.Logf("Tiller Env Setting Failed: %v", err.Error())      
	}           
	helm := NewHelm()
	
        xapp, err := helm.Status(name)
        if err == nil {
                t.Logf("Status returned: %v", err)
        }
        xapp2, err := helm.Delete(name)
        if err != nil {
	 	assert.NotEqual(t, err, "Error: release: not found")		
        }else{
		t.Logf("xapp : %+v, Xapp2 : %+v",xapp,xapp2)
	}
	helm.Init()
	
	if version := helm.GetVersion(name); version != "" {
                t.Logf("GetVersion expected to return empty string, got %v", version)
        }
	
	x := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}
	x.OverrideFile = "../../test/dummy-xapp_values.json"
        if args := helm.GetInstallArgs(x, false); args == "" {
                t.Logf("GetInstallArgs failed: got %v", args)
        }

	if err = os.Setenv("HELMVERSION", "2"); err != nil { 
	        t.Logf("after set Tiller Env Setting Failed: %v", err.Error())      
	}
 
}

func TestAddRepoReturnsErrorIfNoUsernameFile(t *testing.T) {
        if _, err := NewHelm().AddRepo(); err == nil {
                t.Errorf("AddRepo expected to fail but it didn't")
        }
}

func TestAddRepoReturnsErrorIfNoPasswordFile(t *testing.T) {
        defer func() { resetHelmExecMock(); removeTestUsernameFile() }()
        helmExec = mockedHelmExec

        if err := writeTestUsernameFile(); err != nil {
                t.Errorf("AddRepo username file create failed: %s", err)
                return
        }
        if _, err := NewHelm().AddRepo(); err == nil {
                t.Errorf("AddRepo expected to fail but it didn't")
        }
}

func TestInstallSuccess(t *testing.T) {
        name := "dummy-xapp"
        xappDesc := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}

        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetOut = helmStatusOutput

        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetOut = kubeServiceOutput

        xapp, err := NewHelm().Install(xappDesc)
        if err != nil {
                t.Errorf("Install failed: %v", err)
        }
        validateXappModel(t, xapp)
}

func TestInstallReturnsErrorIfHelmRepoUpdateFails(t *testing.T) {
        name := "dummy-xapp"
        xappDesc := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}

        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if _, err := NewHelm().Install(xappDesc); err == nil {
                t.Errorf("Install expected to fail but it didn't")
        }
}

func TestStatusSuccess(t *testing.T) {
        name := "dummy-xapp"

        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetOut = helmStatusOutput

        xapp, err := NewHelm().Status(name)
        if err != nil {
                t.Errorf("Status failed: %v", err)
        }
        validateXappModel(t, xapp)
}

func TestStatusReturnsErrorIfHelmStatusFails(t *testing.T) {
        name := "dummy-xapp"
	
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if _, err := NewHelm().Status(name); err == nil {
                t.Errorf("Status expected to fail but it didn't")
        }else{
	 	assert.Equal(t, err, helmExecRetErr)		
	}
}

func TestParseStatusSuccess(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        var expectedHelmCommand string = ""
        helmExec = mockedHelmExec
        helmExecRetOut = helListOutput

        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetOut = kubeServiceOutput

        xapp, err := NewHelm().ParseStatus("dummy-xapp", helmStatusOutput)
        if err != nil {
                t.Errorf("ParseStatus failed: %v", err)
        }
	
        validateXappModel(t, xapp)

        if cm.EnvHelmVersion == cm.HELM_VERSION_2 {
                expectedHelmCommand = "list --deployed --output yaml --namespace=ricxapp dummy-xapp"
        }else {
                expectedHelmCommand = "list --deployed --output yaml --namespace=ricxapp -f dummy-xapp"
        }
        if caughtHelmExecArgs != expectedHelmCommand {
                t.Errorf("ParseStatus failed: expected %v, got %v", expectedHelmCommand, caughtHelmExecArgs)
        }
}

func TestListSuccess(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetOut = helListAllOutput

        names, err := NewHelm().List()
        if err != nil {
                t.Errorf("List failed: %v", err)
        }

        if !reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
                t.Errorf("List failed: %v", err)
        }
        expectedHelmCommand := "list --all --deployed --output yaml --namespace=ricxapp"
        if caughtHelmExecArgs != expectedHelmCommand {
                t.Errorf("List: expected %v, got %v", expectedHelmCommand, caughtHelmExecArgs)
        }
	
	var str models.AllDeployableXapps
	str = NewHelm().SearchAll()	
	if str != nil {
		t.Logf("Search end..str : %s\n",str)	
	}else{
	 	assert.Nil(t,str)		
	}
}

func TestListReturnsErrorIfHelmListFails(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if _, err := NewHelm().List(); err == nil {
                t.Errorf("List expected to fail but it didn't")
        }

}

func TestDeleteSuccess(t *testing.T) {
        name := "dummy-xapp"

        var expectedHelmCommand string = ""
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetOut = helmStatusOutput

        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetOut = kubeServiceOutput

        xapp, err := NewHelm().Delete(name)
        if err != nil {
                t.Errorf("Delete failed: %v", err)
        }

        validateXappModel(t, xapp)

        if cm.EnvHelmVersion == cm.HELM_VERSION_2 {
                expectedHelmCommand = "del --purge dummy-xapp"
        } else {
                expectedHelmCommand =   "uninstall dummy-xapp -n ricxapp"
        }
        if caughtHelmExecArgs != expectedHelmCommand {
                t.Errorf("Delete failed: expected %v, got %v", expectedHelmCommand, caughtHelmExecArgs)
        }
}

func TestDeleteReturnsErrorIfHelmStatusFails(t *testing.T) {
        name := "dummy-xapp"

        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if _, err := NewHelm().Delete(name); err == nil {
                t.Errorf("Delete expected to fail but it didn't")
        }
}

func TestFetchSuccessIfCmdArgHasTestSuffix(t *testing.T) {
        if err := NewHelm().Fetch("lsfuis", "../../helm_chart/appmgr/values.yaml"); err != nil {
                t.Errorf("Fetch failed: %v", err)
        }
}

func TestGetVersionSuccess(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        var expectedHelmCommand string = ""
        helmExec = mockedHelmExec
        helmExecRetOut = helListOutput

        if version := NewHelm().GetVersion("dummy-xapp"); version != "1.0" {
                t.Errorf("GetVersion failed: expected 1.0, got %v", version)
        }

        if cm.EnvHelmVersion == cm.HELM_VERSION_2{
                expectedHelmCommand = "list --deployed --output yaml --namespace=ricxapp dummy-xapp"
        }else {
                expectedHelmCommand = "list --deployed --output yaml --namespace=ricxapp -f dummy-xapp"
        }
        if caughtHelmExecArgs != expectedHelmCommand {
                t.Errorf("GetVersion failed: expected %v, got %v", expectedHelmCommand, caughtHelmExecArgs)
        }
}

func TestGetVersionReturnsEmptyStringIfHelmListFails(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if version := NewHelm().GetVersion("dummy-xapp"); version != "" {
                t.Errorf("GetVersion expected to return empty string, got %v", version)
        }
}

func TestGetAddressSuccess(t *testing.T) {
        ip, port := NewHelm().GetAddress(helmStatusOutput)
        if ip != "10.102.184.212" {
                t.Errorf("GetAddress failed: expected 10.102.184.212, got %v", ip)
        }
        if port != "80/TCP" {
                t.Errorf("GetAddress failed: expected 80/TCP, got %v", port)
        }
}

func TestGetEndpointInfoSuccess(t *testing.T) {
        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetOut = kubeServiceOutput

        svc, port := NewHelm().GetEndpointInfo("dummy-xapp")
	
        expectedSvc := "service-ricxapp-dummy-xapp-rmr.ricxapp"
        if svc != expectedSvc {
                t.Errorf("GetEndpointInfo failed: expected %v, got %v", expectedSvc, svc)
        }
        if port != 4560 {
                t.Errorf("GetEndpointInfo failed: expected port 4560, got %v", port)
        }
        expectedKubeCommand := " get service -n ricxapp service-ricxapp-dummy-xapp-rmr -o json"
        if caughtKubeExecArgs != expectedKubeCommand {
                t.Errorf("GetEndpointInfo failed: expected %v, got %v", expectedKubeCommand, caughtKubeExecArgs)
        }
}

func TestGetEndpointInfoReturnsDefaultPortIfJsonParseFails(t *testing.T) {
        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetOut = "not-json-syntax"

        svc, port := NewHelm().GetEndpointInfo("dummy-xapp")
        expectedSvc := "service-ricxapp-dummy-xapp-rmr.ricxapp"
        if svc != expectedSvc {
                t.Errorf("GetEndpointInfo failed: expected %v, got %v", expectedSvc, svc)
        }
        if port != 4560 {
                t.Errorf("GetEndpointInfo failed: expected port 4560, got %v", port)
        }
}

func TestGetEndpointInfoReturnsDefaultPortIfKubeGetServiceFails(t *testing.T) {
        defer func() { resetKubeExecMock() }()
        kubeExec = mockedKubeExec
        kubeExecRetErr = errors.New("some helm command error")

        svc, port := NewHelm().GetEndpointInfo("dummy-xapp")
        expectedSvc := "service-ricxapp-dummy-xapp-rmr.ricxapp"
        if svc != expectedSvc {
                t.Errorf("GetEndpointInfo failed: expected %v, got %v", expectedSvc, svc)
        }
        if port != 4560 {
                t.Errorf("GetEndpointInfo failed: expected port 4560, got %v", port)
        }
}

func TestHelmStatusAllSuccess(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetOut = helListAllOutput

        if _, err := NewHelm().StatusAll(); err != nil {
                t.Errorf("StatusAll failed: %v", err)
        }
        // Todo: check StatusAll response content
}

func TestStatusAllReturnsErrorIfHelmListFails(t *testing.T) {
        defer func() { resetHelmExecMock() }()
        helmExec = mockedHelmExec
        helmExecRetErr = errors.New("some helm command error")

        if _, err := NewHelm().StatusAll(); err == nil {
                t.Errorf("StatusAll expected to fail but it didn't")
        }
}

func TestGetNamesSuccess(t *testing.T) {
        names, err := NewHelm().GetNames(helListAllOutput)
        if err != nil {
                t.Errorf("GetNames failed: %v", err)
        }
        if !reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
                t.Errorf("GetNames failed: %v", err)
        }
}

func TestGetNamesFail(t *testing.T) {
        names, err := NewHelm().GetNames("helListAll")
        if err != nil {
                t.Errorf("GetNames failed: %v", err)
        }
        if reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
                t.Errorf("GetNames succ")
        }
}
func TestAddTillerEnv(t *testing.T) {
        if err := NewHelm().AddTillerEnv(); err != nil {
                t.Errorf("AddTillerEnv failed!")
        }
}

func TestGetInstallArgs(t *testing.T) {
        name := "dummy-xapp"
        var expectedArgs string = ""

        x := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}

        if cm.EnvHelmVersion == cm.HELM_VERSION_3 {
                expectedArgs = "install dummy-xapp helm-repo/dummy-xapp --namespace=ricxapp"
        }else {
                expectedArgs = "install helm-repo/dummy-xapp --namespace=ricxapp --name=dummy-xapp"
        }

        if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
                t.Errorf("GetInstallArgs failed: expected '%v', got '%v'", expectedArgs, args)
        }

        expectedArgs += " --set ricapp.appconfig.override=dummy-xapp-appconfig"
        if args := NewHelm().GetInstallArgs(x, true); args != expectedArgs {
                t.Errorf("GetInstallArgs failed: expected %v, got %v", expectedArgs, args)
        }

        x.HelmVersion = "1.2.3"
        if cm.EnvHelmVersion == cm.HELM_VERSION_3 {
                expectedArgs = "install dummy-xapp helm-repo/dummy-xapp --namespace=ricxapp --version=1.2.3"
        } else {
                expectedArgs = "install helm-repo/dummy-xapp --namespace=ricxapp --version=1.2.3 --name=dummy-xapp"
        }
        if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
                t.Errorf("GetInstallArgs failed: expected %v, got %v", expectedArgs, args)
        }


        x.ReleaseName = "ueec-xapp"
        if cm.EnvHelmVersion == cm.HELM_VERSION_3 {
                expectedArgs = "install dummy-xapp helm-repo/dummy-xapp --namespace=ricxapp --version=1.2.3"
        } else {
                expectedArgs = "install helm-repo/dummy-xapp --namespace=ricxapp --version=1.2.3 --name=ueec-xapp"
        }
        if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
                t.Errorf("GetInstallArgs failed: expected %v, got %v", expectedArgs, args)
        }

        x.OverrideFile = "../../test/dummy-xapp_values.json"
        expectedArgs += " -f=/tmp/appmgr_override.yaml"
        if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
                t.Errorf("GetInstallArgs failed: expected %v, got %v", expectedArgs, args)
        }
}

func writeTestUsernameFile() error {
        f, err := os.Create(viper.GetString("helm.helm-username-file"))
        if err != nil {
                return err
        }
        _, err = f.WriteString("some-username")
        f.Close()
        return err
}

func removeTestUsernameFile() error {
        return os.Remove(viper.GetString("helm.helm-username-file"))
}

func writeTestPasswordFile() (err error) {
        f, err := os.Create(viper.GetString("helm.helm-password-file"))
        if err != nil {
                return err
        }

        _, err = f.WriteString("some-password")
        f.Close()
        return err
}

func removeTestPasswordFile() error {
        return os.Remove(viper.GetString("helm.helm-password-file"))
}

func getXappData() (x models.Xapp) {
        //name1 := "dummy-xapp-8984fc9fd-l6xch"
        //name2 := "dummy-xapp-8984fc9fd-pp4hg"
        x = generateXapp("dummy-xapp", "deployed", "1.0", "dummy-xapp-8984fc9fd-bkcbp", "running", "service-ricxapp-dummy-xapp-rmr.ricxapp", "4560")
        //x.Instances = append(x.Instances, x.Instances[0])
        //x.Instances = append(x.Instances, x.Instances[0])
        //x.Instances[1].Name = &name1
        //x.Instances[2].Name = &name2

        return x
}

func generateXapp(name, status, ver, iname, istatus, ip, port string) (x models.Xapp) {
        x.Name = &name
        x.Status = status
        x.Version = ver
        p, _ := strconv.Atoi(port)
        var msgs appmgr.RtmData

        instance := &models.XappInstance{
                Name:       &iname,
                Status:     istatus,
                IP:         ip,
                Port:       int64(p),
                TxMessages: msgs.TxMessages,
                RxMessages: msgs.RxMessages,
        }
        x.Instances = append(x.Instances, instance)

        return
}

func mockedKubeExec(args string) (out []byte, err error) {
        caughtKubeExecArgs = args
        return []byte(kubeExecRetOut), kubeExecRetErr
}

func resetKubeExecMock() {
        kubeExec = util.KubectlExec
        caughtKubeExecArgs = ""
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

func validateXappModel(t *testing.T, xapp models.Xapp) {
        expXapp := getXappData()
        xapp.Version = "1.0"

        if *expXapp.Name != *xapp.Name || expXapp.Status != xapp.Status || expXapp.Version != xapp.Version {
                t.Errorf("\n%v \n%v", *xapp.Name, *expXapp.Name)
        }

        if *expXapp.Instances[0].Name != *xapp.Instances[0].Name || expXapp.Instances[0].Status != xapp.Instances[0].Status {
                t.Errorf("\n1:%v 2:%v", *expXapp.Instances[0].Name, *xapp.Instances[0].Name)
        }

        if expXapp.Instances[0].IP != xapp.Instances[0].IP || expXapp.Instances[0].Port != xapp.Instances[0].Port {
                t.Errorf("\n%v - %v, %v - %v", expXapp.Instances[0].IP, xapp.Instances[0].IP, expXapp.Instances[0].Port, xapp.Instances[0].Port)
        }
}
