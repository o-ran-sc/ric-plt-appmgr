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

package restful

import (
	"net/http"

	cfgmap "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/cm"
	helmer "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/helm"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/restapi/operations"
	resthook "gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/resthooks"
)

type CmdOptions struct {
	hostAddr      *string
	helmHost      *string
	helmChartPath *string
}

type Resource struct {
	Method      string
	Url         string
	HandlerFunc http.HandlerFunc
}

type Restful struct {
	api   *operations.AppManagerAPI
	helm  *helmer.Helm
	cm    *cfgmap.CM
	rh    *resthook.Resthook
	ready bool
}

//Taken from xapp-frame models
type ConfigMetadata struct {

	// The type of the content
	// Required: true
	// Enum: [json xml other]
	ConfigType *string `json:"configType"`

	// Name of the xApp
	// Required: true
	XappName *string `json:"xappName"`
}

type XAppConfig struct {

	// Configuration in JSON format
	// Required: true
	Config interface{} `json:"config"`

	// metadata
	// Required: true
	Metadata *ConfigMetadata `json:"metadata"`
}

type XappConfigList []*XAppConfig
