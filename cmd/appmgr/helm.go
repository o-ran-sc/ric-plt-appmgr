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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var execCommand = exec.Command

func Exec(args string) (out []byte, err error) {
	cmd := execCommand("/bin/sh", "-c", strings.Join([]string{"helm", args}, " "))

	if !strings.HasSuffix(os.Args[0], ".test") {
		out, err = cmd.CombinedOutput()
		if err != nil {
			mdclog(MdclogErr, formatLog("Command failed", args, err.Error()))
		}
		return out, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Running command: %v", cmd)
	for i := 0; i < 3; i++ {
		err = cmd.Run()
		if err != nil {
			mdclog(MdclogErr, formatLog("Command failed, retrying", args, err.Error()+stderr.String()))
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}
		break
	}

	if err == nil && !strings.HasSuffix(os.Args[0], ".test") {
		mdclog(MdclogDebug, formatLog("command success", stdout.String(), ""))
		return stdout.Bytes(), nil
	}

	return stdout.Bytes(), errors.New(stderr.String())
}

func (h *Helm) Initialize() {
	if h.initDone == true {
		return
	}

	for {
		if _, err := h.Init(); err == nil {
			mdclog(MdclogDebug, formatLog("Helm init done successfully!", "", ""))
			break
		}
		mdclog(MdclogErr, formatLog("helm init failed, retyring ...", "", ""))
		time.Sleep(time.Duration(10) * time.Second)
	}

	for {
		if _, err := h.AddRepo(); err == nil {
			mdclog(MdclogDebug, formatLog("Helm repo added successfully", "", ""))
			break
		}
		mdclog(MdclogErr, formatLog("Helm repo addition failed, retyring ...", "", ""))
		time.Sleep(time.Duration(10) * time.Second)
	}
	h.initDone = true
}

func (h *Helm) Run(args string) (out []byte, err error) {
	return Exec(args)
}

// API functions
func (h *Helm) Init() (out []byte, err error) {

	// Add Tiller address as environment variable
	if err := addTillerEnv(); err != nil {
		return out, err
	}

	return Exec(strings.Join([]string{"init -c"}, ""))
}

func (h *Helm) AddRepo() (out []byte, err error) {

	// Get helm repo user name and password from files mounted by secret object
	credFile, err := ioutil.ReadFile(viper.GetString("helm.helm-username-file"))
	if err != nil {
		mdclog(MdclogErr, formatLog("helm_repo_username ReadFile failed", "", err.Error()))
		return
	}

	username := " --username " + string(credFile)

	credFile, err = ioutil.ReadFile(viper.GetString("helm.helm-password-file"))
	if err != nil {
		mdclog(MdclogErr, formatLog("helm_repo_password ReadFile failed", "", err.Error()))
		return
	}

	pwd := " --password " + string(credFile)

	// Get internal helm repo name
	rname := viper.GetString("helm.repo-name")

	// Get helm repo address from values.yaml
	repo := viper.GetString("helm.repo")

	return Exec(strings.Join([]string{"repo add ", rname, " ", repo, username, pwd}, ""))
}

func (h *Helm) Install(name string) (xapp Xapp, err error) {
	out, err := h.Run(strings.Join([]string{"repo update "}, ""))
	if err != nil {
		return
	}

	rname := viper.GetString("helm.repo-name")

	ns := getNamespaceArgs()
	out, err = h.Run(strings.Join([]string{"install ", rname, "/", name, " --name ", name, ns}, ""))
	if err != nil {
		return
	}

	return h.ParseStatus(name, string(out))
}

func (h *Helm) Status(name string) (xapp Xapp, err error) {

	out, err := h.Run(strings.Join([]string{"status ", name}, ""))
	if err != nil {
		mdclog(MdclogErr, formatLog("Getting xapps status", "", err.Error()))
		return
	}

	return h.ParseStatus(name, string(out))
}

func (h *Helm) StatusAll() (xapps []Xapp, err error) {
	xappNameList, err := h.List()
	if err != nil {
		mdclog(MdclogErr, formatLog("Helm list failed", "", err.Error()))
		return
	}

	return h.parseAllStatus(xappNameList)
}

func (h *Helm) List() (names []string, err error) {

	ns := getNamespaceArgs()
	out, err := h.Run(strings.Join([]string{"list --all --output yaml ", ns}, ""))
	if err != nil {
		mdclog(MdclogErr, formatLog("Listing deployed xapps failed", "", err.Error()))
		return
	}

	return h.GetNames(string(out))
}

func (h *Helm) Delete(name string) (xapp Xapp, err error) {
	xapp, err = h.Status(name)
	if err != nil {
		mdclog(MdclogErr, formatLog("Fetching xapp status failed", "", err.Error()))
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
func (h *Helm) GetMessages(name string) (msgs MessageTypes, err error) {
	tarDir := viper.GetString("xapp.tarDir")
	if tarDir == "" {
		tarDir = "/tmp"
	}

	if h.Fetch(name, tarDir); err != nil {
		mdclog(MdclogWarn, formatLog("Fetch chart failed", "", err.Error()))
		return
	}

	return h.ParseMessages(name, tarDir, viper.GetString("xapp.msg_type_file"))

}

func (h *Helm) ParseMessages(name string, chartDir, msgFile string) (msgs MessageTypes, err error) {
	yamlFile, err := ioutil.ReadFile(path.Join(chartDir, name, msgFile))
	if err != nil {
		mdclog(MdclogWarn, formatLog("ReadFile failed", "", err.Error()))
		return
	}

	err = yaml.Unmarshal(yamlFile, &msgs)
	if err != nil {
		mdclog(MdclogWarn, formatLog("Unmarshal failed", "", err.Error()))
		return
	}

	if err = os.RemoveAll(path.Join(chartDir, name)); err != nil {
		mdclog(MdclogWarn, formatLog("RemoveAll failed", "", err.Error()))
	}

	return
}

func (h *Helm) GetVersion(name string) (version string) {

	ns := getNamespaceArgs()
	out, err := h.Run(strings.Join([]string{"list --output yaml ", name, ns}, ""))
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
	r := regexp.MustCompile(`(?s)\/Pod.*?\/Service`)
	result := r.FindStringSubmatch(string(out))
	if result == nil {
		return
	}

	re := regexp.MustCompile(name + "-(\\d+).*")
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

	types, err := h.GetMessages(name)
	if err != nil {
		// xAPP can still be deployed if the msg_type file is missing.
		mdclog(MdclogWarn, formatLog("method GetMessages Failed....", "", err.Error()))

		//Set err back to nil, so it does not cause issues in called functions.
		err = nil
	}

	h.FillInstanceData(name, out, &xapp, types)

	return
}

func (h *Helm) parseAllStatus(names []string) (xapps []Xapp, err error) {
	xapps = []Xapp{}

	for _, name := range names {
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
		mdclog(MdclogErr, formatLog("Tiller Env Setting Failed", "", err.Error()))
	}

	return err
}

func getNamespaceArgs() string {
	ns := viper.GetString("xapp.namespace")
	if ns == "" {
		ns = "ricxapp"
	}
	return " --namespace=" + ns
}

func formatLog(text string, args string, err string) string {
	return fmt.Sprintf("Helm: %s: args=%s err=%s\n", text, args, err)
}
