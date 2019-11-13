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
	"os"
	"reflect"
	"strconv"
	"testing"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/util"
)

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

var helListOutput = `Next: ""
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
  Updated: Sun Mar 24 07:17:00 2019`

// Test cases
func TestMain(m *testing.M) {
	appmgr.Init()
	appmgr.Logger.SetLevel(0)

	code := m.Run()
	os.Exit(code)
}

func TestHelmStatus(t *testing.T) {
	//NewHelm().SetCM(&ConfigMap{})
	util.KubectlExec = func(args string) (out []byte, err error) {
		return []byte("10.102.184.212"), nil
	}
	xapp, err := NewHelm().ParseStatus("dummy-xapp", helmStatusOutput)
	if err != nil {
		t.Errorf("Helm install failed: %v", err)
	}
	x := getXappData()
	xapp.Version = "1.0"

	if *x.Name != *xapp.Name || x.Status != xapp.Status || x.Version != xapp.Version {
		t.Errorf("\n%v \n%v", *xapp.Name, *x.Name)
	}

	if *x.Instances[0].Name != *xapp.Instances[0].Name || x.Instances[0].Status != xapp.Instances[0].Status {
		t.Errorf("\n1:%v 2:%v", *x.Instances[0].Name, *xapp.Instances[0].Name)
	}

	if x.Instances[0].IP != xapp.Instances[0].IP || x.Instances[0].Port != xapp.Instances[0].Port {
		t.Errorf("\n1:%v 2:%v", x.Instances[0].IP, xapp.Instances[0].IP)
	}
}

func TestHelmLists(t *testing.T) {
	names, err := NewHelm().GetNames(helListOutput)
	if err != nil {
		t.Errorf("Helm status failed: %v", err)
	}

	if !reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
		t.Errorf("Helm status failed: %v", err)
	}
}

func TestAddTillerEnv(t *testing.T) {
	if NewHelm().AddTillerEnv() != nil {
		t.Errorf("TestAddTillerEnv failed!")
	}
}

func TestGetInstallArgs(t *testing.T) {
	name := "dummy-xapp"
	x := models.XappDescriptor{XappName: &name, Namespace: "ricxapp"}

	expectedArgs := "install helm-repo/dummy-xapp  --namespace=ricxapp --name=dummy-xapp"
	if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
		t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}

	x.HelmVersion = "1.2.3"
	expectedArgs = "install helm-repo/dummy-xapp  --namespace=ricxapp --version=1.2.3 --name=dummy-xapp"
	if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
		t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}

	x.ReleaseName = "ueec-xapp"
	expectedArgs = "install helm-repo/dummy-xapp  --namespace=ricxapp --version=1.2.3 --name=ueec-xapp"
	if args := NewHelm().GetInstallArgs(x, false); args != expectedArgs {
		t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}
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
	var msgs appmgr.MessageTypes

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
