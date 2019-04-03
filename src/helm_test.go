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
    "fmt"
    "os"
    "os/exec"
	"strconv"
    "testing"
    "reflect"
    "github.com/spf13/viper"
    "io/ioutil"
    "path"
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

var mockedExitStatus = 0
var mockedStdout string
var h = Helm{}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
    cs := []string{"-test.run=TestExecCommandHelper", "--", command}
    cs = append(cs, args...)
	
	cmd := exec.Command(os.Args[0], cs...)
    es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "STDOUT=" + mockedStdout, "EXIT_STATUS=" + es}
	
    return cmd
}

func TestExecCommandHelper(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
        return
    }

    fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
    i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
    os.Exit(i)
}

func writeTestCreds() (err error) {

    // Write test entries to helm username and password files
    f, err := os.Create(viper.GetString("helm.helm-username-file"))
    if err != nil {
        return err
    }

    _, err = f.WriteString(viper.GetString("helm.secrets.username"))
    if err != nil {
        f.Close()
        return (err)
    }
    f.Close()

    f, err = os.Create(viper.GetString("helm.helm-password-file"))
    if err != nil {
        return err
    }

    _, err = f.WriteString(viper.GetString("helm.secrets.password"))
    if err != nil {
        f.Close()
        return (err)
    }
    f.Close()
    return
}

func TestHelmInit(t *testing.T) {
	mockedExitStatus = 0
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    if err := writeTestCreds(); err != nil {
        t.Errorf("Writing test entries failed: %s", err)
        return
    }

    out, err := h.Init()
    if err != nil {
        t.Errorf("Helm init failed: %s %s", err, string(out))
    }
}

func TestHelmInstall(t *testing.T) {
    copyFile(t)
    mockedExitStatus = 0
	execCommand = fakeExecCommand
	mockedStdout = helmStatusOutput
    defer func() { execCommand = exec.Command }()

    xapp, err := h.Install("dummy-xapp")
    if err != nil {
        t.Errorf("Helm install failed: %v", err)
	}

    x := getXappData()
    xapp.Version = "1.0"

    if !reflect.DeepEqual(xapp, x) {
        t.Errorf("%v \n%v", xapp, x)
    }
}

func TestHelmStatus(t *testing.T) {
    copyFile(t)
    mockedExitStatus = 0
    mockedStdout = helmStatusOutput
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    xapp, err := h.Status("dummy-xapp")
    if err != nil {
        t.Errorf("Helm status failed: %v", err)
	}

    x := getXappData()
    xapp.Version = "1.0"

	if !reflect.DeepEqual(xapp, x) {
        t.Errorf("%v \n%v", xapp, x)
    }
}

func TestHelmStatusAll(t *testing.T) {
    copyFile(t)
    mockedExitStatus = 0
    mockedStdout = helListOutput
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    xapp, err := h.StatusAll()
    if err != nil {
        t.Errorf("Helm StatusAll failed: %v - %v", err, xapp)
	}

    // Todo: check the content
}

func TestHelmParseAllStatus(t *testing.T) {
    copyFile(t)
    mockedExitStatus = 0
    mockedStdout = helListOutput
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    xapp, err := h.parseAllStatus([]string{"dummy-xapp", "dummy-xapp2"})
    if err != nil {
        t.Errorf("Helm parseAllStatus failed: %v - %v", err, xapp)
	}

    // Todo: check the content
}

func TestHelmDelete(t *testing.T) {
    copyFile(t)
    mockedExitStatus = 0
    mockedStdout = helListOutput
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    xapp, err := h.Delete("dummy-xapp")
    if err != nil {
        t.Errorf("Helm delete failed: %v - %v", err, xapp)
	}

    // Todo: check the content
}

func TestHelmLists(t *testing.T) {
    mockedExitStatus = 0
    mockedStdout = helListOutput
    execCommand = fakeExecCommand
    defer func() { execCommand = exec.Command }()

    names, err := h.List()
    if err != nil {
        t.Errorf("Helm status failed: %v", err)
	}

    if !reflect.DeepEqual(names, []string{"dummy-xapp", "dummy-xapp2"}) {
        t.Errorf("Helm status failed: %v", err)
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


func copyFile(t *testing.T) {
    tarDir := path.Join(viper.GetString("xapp.tarDir"), "dummy-xapp")
    err := os.MkdirAll(tarDir, 0777)
    if err != nil {
         t.Errorf("%v", err)
    }

    data, err := ioutil.ReadFile("../config/msg_type.yaml")
    if err != nil {
         t.Errorf("%v", err)
    }

    _ = ioutil.WriteFile(path.Join(tarDir, "msg_type.yaml"), data, 0644)
    if err != nil {
         t.Errorf("%v", err)
    }
}