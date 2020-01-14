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
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/cm"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/util"
)

type Helm struct {
	initDone bool
	cm       *cm.CM
}

func NewHelm() *Helm {
	return &Helm{initDone: false, cm: cm.NewCM()}
}

func (h *Helm) Initialize() {
	if h.initDone == true {
		return
	}

	for {
		if _, err := h.Init(); err == nil {
			appmgr.Logger.Info("Helm init done successfully!")
			break
		}
		appmgr.Logger.Info("helm init failed, retyring ...")
		time.Sleep(time.Duration(10) * time.Second)
	}

	for {
		if _, err := h.AddRepo(); err == nil {
			appmgr.Logger.Info("Helm repo added successfully")
			break
		}
		appmgr.Logger.Info("Helm repo addition failed, retyring ...")
		time.Sleep(time.Duration(10) * time.Second)
	}
	h.initDone = true
}

func (h *Helm) Run(args string) (out []byte, err error) {
	return util.HelmExec(args)
}

// API functions
func (h *Helm) Init() (out []byte, err error) {
	if err := h.AddTillerEnv(); err != nil {
		return out, err
	}

	return util.HelmExec(strings.Join([]string{"init -c --skip-refresh"}, ""))
}

func (h *Helm) AddRepo() (out []byte, err error) {
	// Get helm repo user name and password from files mounted by secret object
	credFile, err := ioutil.ReadFile(viper.GetString("helm.helm-username-file"))
	if err != nil {
		appmgr.Logger.Info("helm_repo_username ReadFile failed: %v", err.Error())
		return
	}
	username := " --username " + string(credFile)

	credFile, err = ioutil.ReadFile(viper.GetString("helm.helm-password-file"))
	if err != nil {
		appmgr.Logger.Info("helm_repo_password ReadFile failed: %v", err.Error())
		return
	}
	pwd := " --password " + string(credFile)

	rname := viper.GetString("helm.repo-name")
	repo := viper.GetString("helm.repo")

	return util.HelmExec(strings.Join([]string{"repo add ", rname, " ", repo, username, pwd}, ""))
}

func (h *Helm) Install(m models.XappDescriptor) (xapp models.Xapp, err error) {
	m.Namespace = h.cm.GetNamespace(m.Namespace)

	out, err := h.Run(strings.Join([]string{"repo update "}, ""))
	if err != nil {
		return
	}

	out, err = h.Run(h.GetInstallArgs(m, false))
	if err != nil {
		return
	}
	return h.ParseStatus(*m.XappName, string(out))
}

func (h *Helm) Status(name string) (xapp models.Xapp, err error) {
	out, err := h.Run(strings.Join([]string{"status ", name}, ""))
	if err != nil {
		appmgr.Logger.Info("Getting xapps status: %v", err.Error())
		return
	}

	return h.ParseStatus(name, string(out))
}

func (h *Helm) StatusAll() (xapps models.AllDeployedXapps, err error) {
	xappNameList, err := h.List()
	if err != nil {
		appmgr.Logger.Info("Helm list failed: %v", err.Error())
		return
	}

	return h.parseAllStatus(xappNameList)
}

func (h *Helm) List() (names []string, err error) {
	ns := h.cm.GetNamespace("")
	out, err := h.Run(strings.Join([]string{"list --all --deployed --output yaml --namespace=", ns}, ""))
	if err != nil {
		appmgr.Logger.Info("Listing deployed xapps failed: %v", err.Error())
		return
	}

	return h.GetNames(string(out))
}

func (h *Helm) SearchAll() models.AllDeployableXapps {
	return h.cm.GetNamesFromHelmRepo()
}

func (h *Helm) Delete(name string) (xapp models.Xapp, err error) {
	xapp, err = h.Status(name)
	if err != nil {
		appmgr.Logger.Info("Fetching xapp status failed: %v", err.Error())
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
	out, err := h.Run(strings.Join([]string{"list --deployed --output yaml --namespace=", ns, " ", name}, ""))
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

func (h *Helm) GetEndpointInfo(name string) (ip string, port int) {
	ns := h.cm.GetNamespace("")
	args := fmt.Sprintf(" get endpoints -o=jsonpath='{.subsets[*].addresses[*].ip}' service-%s-%s-rmr -n %s", ns, name, ns)
	out, err := util.KubectlExec(args)
	if err != nil {
		return
	}
	appmgr.Logger.Info("Endpoint IP address of %s: %s", name, string(out))
	return fmt.Sprintf("service-%s-%s-rmr.%s", ns, name, ns), 4560
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

func (h *Helm) FillInstanceData(name string, out string, xapp *models.Xapp, rtData appmgr.RtmData) {
	ip, port := h.GetEndpointInfo(name)
	if ip == "" {
		appmgr.Logger.Info("Endpoint IP address not found, using CluserIP")
		ip, _ = h.GetAddress(out)
	}

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
			var x models.XappInstance
			var name string
			fmt.Sscanf(v[0], "%s %s %s", &name, &tmp, &x.Status)
			x.Name = &name
			x.Status = strings.ToLower(x.Status)
			x.IP = ip
			x.Port = int64(port)
			x.TxMessages = rtData.TxMessages
			x.RxMessages = rtData.RxMessages
			x.Policies = rtData.Policies
			xapp.Instances = append(xapp.Instances, &x)
		}
	}
}

func (h *Helm) ParseStatus(name string, out string) (xapp models.Xapp, err error) {
	xapp.Name = &name
	xapp.Version = h.GetVersion(name)
	xapp.Status = h.GetState(out)

	h.FillInstanceData(name, out, &xapp, h.cm.GetRtmData(name))
	return
}

func (h *Helm) parseAllStatus(names []string) (xapps models.AllDeployedXapps, err error) {
	xapps = models.AllDeployedXapps{}
	for _, name := range names {
		var desc interface{}
		err := h.cm.ReadSchema(name, &desc)
		if err != nil {
			continue
		}

		x, err := h.Status(name)
		if err == nil {
			xapps = append(xapps, &x)
		}
	}
	return
}

func (h *Helm) AddTillerEnv() (err error) {
	service := viper.GetString("helm.tiller-service")
	namespace := viper.GetString("helm.tiller-namespace")
	port := viper.GetString("helm.tiller-port")

	if err = os.Setenv("HELM_HOST", service+"."+namespace+":"+port); err != nil {
		appmgr.Logger.Info("Tiller Env Setting Failed: %v", err.Error())
	}
	return err
}

func (h *Helm) GetInstallArgs(x models.XappDescriptor, cmOverride bool) (args string) {
	args = fmt.Sprintf("%s --namespace=%s", args, x.Namespace)
	if x.HelmVersion != "" {
		args = fmt.Sprintf("%s --version=%s", args, x.HelmVersion)
	}

	if x.ReleaseName != "" {
		args = fmt.Sprintf("%s --name=%s", args, x.ReleaseName)
	} else {
		args = fmt.Sprintf("%s --name=%s", args, *x.XappName)
	}

	if cmOverride == true {
		args = fmt.Sprintf("%s ---set ricapp.appconfig.override=%s-appconfig", args, *x.XappName)
	}

	if x.OverrideFile != nil {
		if overrideYaml, err := yaml.JSONToYAML([]byte(x.OverrideFile.(string))); err == nil {
			err = ioutil.WriteFile("/tmp/appmgr_override.yaml", overrideYaml, 0644)
			if err != nil {
				appmgr.Logger.Info("ioutil.WriteFile(/tmp/appmgr_override.yaml) failed: %v", err)
			} else {
				args = args + " -f=/tmp/appmgr_override.yaml"
			}
		} else {
			appmgr.Logger.Info("yaml.JSONToYAML failed: %v", err)
		}
	}

	repoName := viper.GetString("helm.repo-name")
	if repoName == "" {
		repoName = "helm-repo"
	}
	return fmt.Sprintf("install %s/%s %s", repoName, *x.XappName, args)
}
