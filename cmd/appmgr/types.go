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
	"github.com/gorilla/mux"
	cmap "github.com/orcaman/concurrent-map"
	"net/http"
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

type Xapp struct {
	Name       string         `json:"name"`
	ConfigName string         `json:"configName, omitempty"`
	Namespace  string         `json:"namespace, omitempty"`
	Status     string         `json:"status"`
	Version    string         `json:"version"`
	Instances  []XappInstance `json:"instances"`
}

type XappInstance struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Ip         string   `json:"ip"`
	Port       int      `json:"port"`
	TxMessages []string `json:"txMessages"`
	RxMessages []string `json:"rxMessages"`
}

type XappManager struct {
	router *mux.Router
	helm   Helmer
	sd     SubscriptionDispatcher
	opts   CmdOptions
	ready  bool
}

type Helmer interface {
	Initialize()
	Install(m ConfigMetadata) (xapp Xapp, err error)
	Status(name string) (xapp Xapp, err error)
	StatusAll() (xapps []Xapp, err error)
	List() (xapps []string, err error)
	Delete(name string) (xapp Xapp, err error)
}

type Helm struct {
	host      string
	chartPath string
	initDone  bool
}

type SubscriptionReq struct {
	Id         string `json:"id"`
	TargetUrl  string `json:"targetUrl"`
	EventType  string `json:"eventType"`
	MaxRetries int    `json:"maxRetries"`
	RetryTimer int    `json:"retryTimer"`
}

type SubscriptionResp struct {
	Id        string `json:"id"`
	Version   int    `json:"version"`
	EventType string `json:"eventType"`
}

type SubscriptionNotif struct {
	Id        string `json:"id"`
	Version   int    `json:"version"`
	EventType string `json:"eventType"`
	XappData  []Xapp `json:"xapp"`
}

type Subscription struct {
	req  SubscriptionReq
	resp SubscriptionResp
}

type SubscriptionDispatcher struct {
	client        *http.Client
	subscriptions cmap.ConcurrentMap
	db            *DB
	Seq           int
}

type MessageTypes struct {
	TxMessages []string `yaml:"txMessages"`
	RxMessages []string `yaml:"rxMessages"`
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
