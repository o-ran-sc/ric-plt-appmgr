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
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var execCommand = exec.Command

func Exec(args string) (out []byte, err error) {
	cmd := execCommand("/bin/sh", "-c", args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Println("Running command: ", cmd)
	for i := 0; i < viper.GetInt("helm.retry"); i++ {
		err = cmd.Run()
		if err != nil {
			Logger.Error("Command '%s' failed with error: %v, retrying", args, err.Error()+stderr.String())
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		break
	}

	if err == nil && !strings.HasSuffix(os.Args[0], ".test") {
		Logger.Info("command success: %s", stdout.String())
		return stdout.Bytes(), nil
	}

	return stdout.Bytes(), errors.New(stderr.String())
}

var HelmExec = func(args string) (out []byte, err error) {
	return Exec(strings.Join([]string{"helm", args}, " "))
}

var KubectlExec = func(args string) (out []byte, err error) {
	return Exec(strings.Join([]string{"kubectl", args}, " "))
}

func (h *Helm) SetCM(cm ConfigMapper) {
	h.cm = cm
}

func (h *Helm) Initialize() {
	if h.initDone == true {
		return
	}

	for {
		if _, err := h.Init(); err == nil {
			Logger.Info("Helm init done successfully!")
			break
		}
		Logger.Error("helm init failed, retyring ...")
		time.Sleep(time.Duration(10) * time.Second)
	}

	for {
		if _, err := h.AddRepo(); err == nil {
			Logger.Info("Helm repo added successfully")
			break
		}
		Logger.Error("Helm repo addition failed, retyring ...")
		time.Sleep(time.Duration(10) * time.Second)
	}

	h.initDone = true
}

func (h *Helm) Run(args string) (out []byte, err error) {
	return HelmExec(args)
}

// API functions
func (h *Helm) Init() (out []byte, err error) {
	// Add Tiller address as environment variable
	if err := addTillerEnv(); err != nil {
		return out, err
	}

	return HelmExec(strings.Join([]string{"init -c"}, ""))
}

func (h *Helm) AddRepo() (out []byte, err error) {
	// Get helm repo user name and password from files mounted by secret object
	credFile, err := ioutil.ReadFile(viper.GetString("helm.helm-username-file"))
	if err != nil {
		Logger.Error("helm_repo_username ReadFile failed: %v", err.Error())
		return
	}

	username := " --username " + string(credFile)

	credFile, err = ioutil.ReadFile(viper.GetString("helm.helm-password-file"))
	if err != nil {
		Logger.Error("helm_repo_password ReadFile failed: %v", err.Error())
		return
	}

	pwd := " --password " + string(credFile)

	// Get internal helm repo name
	rname := viper.GetString("helm.repo-name")

	// Get helm repo address from values.yaml
	repo := viper.GetString("helm.repo")

	return HelmExec(strings.Join([]string{"repo add ", rname, " ", repo, username, pwd}, ""))
}

func (h *Helm) Install(m XappDeploy) (xapp Xapp, err error) {
	out, err := h.Run(strings.Join([]string{"repo update "}, ""))
	if err != nil {
		return
	}

	var cm interface{}
	m.Namespace = h.cm.GetNamespace(m.Namespace)

	if err = h.cm.GetConfigMap(m, &cm); err != nil {
		out, err = h.Run(getInstallArgs(m, false))
		if err != nil {
			return
		}
		return h.ParseStatus(m.Name, string(out))
	}

	// ConfigMap exists, try to override
	out, err = h.Run(getInstallArgs(m, true))
	if err == nil {
		return h.ParseStatus(m.Name, string(out))
	}

	cm, cmErr := h.cm.PurgeConfigMap(m)
	out, err = h.Run(getInstallArgs(m, false))
	if err != nil {
		return
	}

	if cmErr == nil {
		cmErr = h.cm.RestoreConfigMap(m, cm)
	}
	return h.ParseStatus(m.Name, string(out))
}

func (h *Helm) Status(name string) (xapp Xapp, err error) {
	out, err := h.Run(strings.Join([]string{"status ", name}, ""))
	if err != nil {
		Logger.Error("Getting xapps status: %v", err.Error())
		return
	}

	return h.ParseStatus(name, string(out))
}

func (h *Helm) StatusAll() (xapps []Xapp, err error) {
	xappNameList, err := h.List()
	if err != nil {
		Logger.Error("Helm list failed: %v", err.Error())
		return
	}

	return h.parseAllStatus(xappNameList)
}

func (h *Helm) List() (names []string, err error) {
	ns := h.cm.GetNamespace("")
	out, err := h.Run(strings.Join([]string{"list --all --output yaml --namespace=", ns}, ""))
	if err != nil {
		Logger.Error("Listing deployed xapps failed: %v", err.Error())
		return
	}

	return h.GetNames(string(out))
}

func (h *Helm) SearchAll() (names []string) {
	return h.cm.GetNamesFromHelmRepo()
}

func (h *Helm) Delete(name string) (xapp Xapp, err error) {
	xapp, err = h.Status(name)
	if err != nil {
		Logger.Error("Fetching xapp status failed: %v", err.Error())
		return
	}

	_, err = h.Run(strings.Join([]string{"del --purge ", name}, ""))
	return xapp, err
}

func (h *Helm) Fetch(name, tarDir string) error {
	if strings.HasSuffix(os.Args[0], ".test") {
		return nil
	}

	rname := viper.GetString("helm.repo-name") + "/"

	_, err := h.Run(strings.Join([]string{"fetch --untar --untardir ", tarDir, " ", rname, name}, ""))
	return err
}

// Helper functions
func (h *Helm) GetVersion(name string) (version string) {
	ns := h.cm.GetNamespace("")
	out, err := h.Run(strings.Join([]string{"list --output yaml --namespace=", ns, " ", name}, ""))
	if err != nil {
		return
	}

	var re = regexp.MustCompile(`AppVersion: .*`)
	ver := re.FindStringSubmatch(string(out))
	if ver != nil {
		version = strings.Split(ver[0], ": ")[1]
		version, _ = strconv.Unquote(version)
	}

	return
}

func (h *Helm) GetState(out string) (status string) {
	re := regexp.MustCompile(`STATUS: .*`)
	result := re.FindStringSubmatch(string(out))
	if result != nil {
		status = strings.ToLower(strings.Split(result[0], ": ")[1])
	}

	return
}

func (h *Helm) GetAddress(out string) (ip, port string) {
	var tmp string
	re := regexp.MustCompile(`ClusterIP.*`)
	addr := re.FindStringSubmatch(string(out))
	if addr != nil {
		fmt.Sscanf(addr[0], "%s %s %s %s", &tmp, &ip, &tmp, &port)
	}

	return
}

func (h *Helm) GetNames(out string) (names []string, err error) {
	re := regexp.MustCompile(`Name: .*`)
	result := re.FindAllStringSubmatch(out, -1)
	if result == nil {
		return
	}

	for _, v := range result {
		xappName := strings.Split(v[0], ": ")[1]
		if strings.Contains(xappName, "appmgr") == false {
			names = append(names, xappName)
		}
	}
	return names, nil
}

func (h *Helm) FillInstanceData(name string, out string, xapp *Xapp, msgs MessageTypes) {
	ip, port := h.GetAddress(out)

	var tmp string
	r := regexp.MustCompile(`.*(?s)(Running|Pending|Succeeded|Failed|Unknown).*?\r?\n\r?\n`)
	result := r.FindStringSubmatch(string(out))
	if result == nil {
		return
	}

	re := regexp.MustCompile(name + "-(\\w+-\\w+).*")
	resources := re.FindAllStringSubmatch(string(result[0]), -1)
	if resources != nil {
		for _, v := range resources {
			var x XappInstance
			fmt.Sscanf(v[0], "%s %s %s", &x.Name, &tmp, &x.Status)
			x.Status = strings.ToLower(x.Status)
			x.Ip = ip
			x.Port, _ = strconv.Atoi(strings.Split(port, "/")[0])
			x.TxMessages = msgs.TxMessages
			x.RxMessages = msgs.RxMessages
			xapp.Instances = append(xapp.Instances, x)
		}
	}
}

func (h *Helm) ParseStatus(name string, out string) (xapp Xapp, err error) {
	xapp.Name = name
	xapp.Version = h.GetVersion(name)
	xapp.Status = h.GetState(out)

	h.FillInstanceData(name, out, &xapp, h.cm.GetMessages(name))

	return
}

func (h *Helm) parseAllStatus(names []string) (xapps []Xapp, err error) {
	xapps = []Xapp{}

	for _, name := range names {
		err := h.cm.ReadSchema(name, &XAppConfig{})
		if err != nil {
			continue
		}

		x, err := h.Status(name)
		if err == nil {
			xapps = append(xapps, x)
		}
	}

	return
}

func addTillerEnv() (err error) {
	service := viper.GetString("helm.tiller-service")
	namespace := viper.GetString("helm.tiller-namespace")
	port := viper.GetString("helm.tiller-port")

	if err = os.Setenv("HELM_HOST", service+"."+namespace+":"+port); err != nil {
		Logger.Error("Tiller Env Setting Failed: %v", err.Error())
	}

	return err
}

func getInstallArgs(x XappDeploy, cmOverride bool) (args string) {
	args = args + " --namespace=" + x.Namespace

	if x.ImageRepo != "" {
		args = args + " --set global.repository=" + x.ImageRepo
	}

	if x.ServiceName != "" {
		args = args + " --set ricapp.service.name=" + x.ServiceName
	}

	if x.Hostname != "" {
		args = args + " --set ricapp.hostname=" + x.Hostname
	}

	if cmOverride == true {
		args = args + " --set ricapp.appconfig.override=" + x.Name + "-appconfig"
	}

	rname := viper.GetString("helm.repo-name")
	return fmt.Sprintf("install %s/%s --name=%s %s", rname, x.Name, x.Name, args)
}
