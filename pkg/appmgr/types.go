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

package appmgr

import (
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
)

type ConfigMap struct {
	Kind       string      `json:"kind"`
	ApiVersion string      `json:"apiVersion"`
	Data       interface{} `json:"data"`
	Metadata   CMMetadata  `json:"metadata"`
}

type CMMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ConfigMapper interface {
	UploadConfig() (cfg []models.XAppConfig)
	GetConfigMap(m models.XappDescriptor, c *interface{}) (err error)
	CreateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error)
	UpdateConfigMap(r models.XAppConfig) (errList models.ConfigValidationErrors, err error)
	DeleteConfigMap(r models.XAppConfig) (cm interface{}, err error)
	ReadSchema(name string, c *models.XAppConfig) (err error)
	PurgeConfigMap(m models.XappDescriptor) (cm interface{}, err error)
	RestoreConfigMap(m models.XappDescriptor, cm interface{}) (err error)
	ReadConfigMap(name string, ns string, c *interface{}) (err error)
	ApplyConfigMap(r models.XAppConfig, action string) (err error)
	GetRtmData(name string) (msgs RtmData)
	GetNamespace(ns string) string
	GetNamesFromHelmRepo() (names []string)
}

type Helmer interface {
	SetCM(ConfigMapper)
	Initialize()
	Install(m models.XappDescriptor) (xapp models.Xapp, err error)
	Status(name string) (xapp models.Xapp, err error)
	StatusAll() (xapps models.AllDeployedXapps, err error)
	SearchAll() (xapps []string)
	List() (xapps []string, err error)
	Delete(name string) (xapp models.Xapp, err error)
}

type Helm struct {
	host      string
	chartPath string
	initDone  bool
	cm        ConfigMapper
}

type RtmData struct {
	TxMessages []string `json:"txMessages"`
	RxMessages []string `json:"rxMessages"`
	Policies   []int64  `json:"Policies"`
}

type EventType string

const (
	Created EventType = "created"
	Updated EventType = "updated"
	Deleted EventType = "deleted"
)

const (
	MdclogErr   = 1 //! Error level log entry
	MdclogWarn  = 2 //! Warning level log entry
	MdclogInfo  = 3 //! Info level log entry
	MdclogDebug = 4 //! Debug level log entry
)
