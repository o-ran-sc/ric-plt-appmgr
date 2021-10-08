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

package resthooks

import (
	cmap "github.com/orcaman/concurrent-map"
	"net/http"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
)

type SubscriptionInfo struct {
	Id   string
	req  models.SubscriptionRequest
	resp models.SubscriptionResponse
}

type Resthook struct {
	client        *http.Client
	subscriptions cmap.ConcurrentMap
	db            iSdl
	Seq           int64
}

// TODO: remove this when RTMGR changes done
type SubscriptionNotification struct {
	Event   string `json:"eventType,omitempty"`
	ID      string `json:"id,omitempty"`
	Version int64  `json:"version,omitempty"`
	XApps   string `json:"xApps,omitempty"`
}

type iSdl interface {
	Set(ns string, pairs ...interface{}) error
	Get(ns string, keys []string) (map[string]interface{}, error)
	GetAll(ns string) ([]string, error)
	RemoveAll(ns string) error
	Remove(ns string, keys []string) error
}
