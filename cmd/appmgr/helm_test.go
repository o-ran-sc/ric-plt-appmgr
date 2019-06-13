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


var h = Helm{}

func TestHelmStatus(t *testing.T) {
	h.SetCM(&ConfigMap{})
	xapp, err := h.ParseStatus("dummy-xapp", helmStatusOutput)
    if err != nil {
        t.Errorf("Helm install failed: %v", err)
	}

    x := getXappData()
    xapp.Version = "1.0"

    if !reflect.DeepEqual(xapp, x) {
        t.Errorf("\n%v \n%v", xapp, x)
    }
}

func TestHelmLists(t *testing.T) {
    names, err := h.GetNames(helListOutput)
    if err != nil {
        t.Errorf("Helm status failed: %v", err)
	}

    if !reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
        t.Errorf("Helm status failed: %v", err)
    }
}

func TestAddTillerEnv(t *testing.T) {
    if addTillerEnv() != nil {
        t.Errorf("TestAddTillerEnv failed!")
	}
}

func TestGetInstallArgs(t *testing.T) {
	x := XappDeploy{Name: "dummy-xapp", Namespace: "ricxapp"}

	expectedArgs := "install helm-repo/dummy-xapp --name=dummy-xapp  --namespace=ricxapp"
	if args := getInstallArgs(x, false); args != expectedArgs {
        t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}

	x.ImageRepo = "localhost:5000"
	expectedArgs = expectedArgs + " --set global.repository=" + "localhost:5000"
	if args := getInstallArgs(x, false); args != expectedArgs {
        t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}

	x.ServiceName = "xapp"
	expectedArgs = expectedArgs + " --set ricapp.service.name=" + "xapp"
	if args := getInstallArgs(x, false); args != expectedArgs {
        t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}

	x.ServiceName = "xapp"
	expectedArgs = expectedArgs + " --set ricapp.appconfig.override=dummy-xapp-appconfig"
	if args := getInstallArgs(x, true); args != expectedArgs {
        t.Errorf("TestGetInstallArgs failed: expected %v, got %v", expectedArgs, args)
	}
}

func getXappData() (x Xapp) {
    x = generateXapp("dummy-xapp", "deployed", "1.0", "dummy-xapp-8984fc9fd-bkcbp", "running", "10.102.184.212", "80")
    x.Instances = append(x.Instances, x.Instances[0])
    x.Instances = append(x.Instances, x.Instances[0])
    x.Instances[1].Name = "dummy-xapp-8984fc9fd-l6xch"
    x.Instances[2].Name = "dummy-xapp-8984fc9fd-pp4hg"

    return x
}

