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
	"encoding/json"
	"fmt"
	"github.com/segmentio/ksuid"
	"net/http"
	"time"
)

func (sd *SubscriptionDispatcher) Initialize() {
	sd.client = &http.Client{}

	sd.db = &DB{}
	sd.db.Create()
	sd.subscriptions = sd.db.RestoreSubscriptions()
}

func (sd *SubscriptionDispatcher) Add(sr SubscriptionReq) (resp SubscriptionResp) {
	key := ksuid.New().String()
	resp = SubscriptionResp{key, 0, sr.EventType}
	sr.Id = key

	sd.subscriptions.Set(key, Subscription{sr, resp})
	sd.db.StoreSubscriptions(sd.subscriptions)

	mdclog(MdclogInfo, fmt.Sprintf("Sub: New subscription added: key=%s value=%v", key, sr))
	return
}

func (sd *SubscriptionDispatcher) GetAll() (hooks []SubscriptionReq) {
	hooks = []SubscriptionReq{}
	for v := range sd.subscriptions.IterBuffered() {
		hooks = append(hooks, v.Val.(Subscription).req)
	}

	return hooks
}

func (sd *SubscriptionDispatcher) Get(id string) (SubscriptionReq, bool) {
	if v, found := sd.subscriptions.Get(id); found {
		mdclog(MdclogInfo, fmt.Sprintf("Subscription id=%s found: %v", id, v.(Subscription).req))

		return v.(Subscription).req, found
	}
	return SubscriptionReq{}, false
}

func (sd *SubscriptionDispatcher) Delete(id string) (SubscriptionReq, bool) {
	if v, found := sd.subscriptions.Get(id); found {
		mdclog(MdclogInfo, fmt.Sprintf("Subscription id=%s found: %v ... deleting", id, v.(Subscription).req))

		sd.subscriptions.Remove(id)
		sd.db.StoreSubscriptions(sd.subscriptions)

		return v.(Subscription).req, found
	}
	return SubscriptionReq{}, false
}

func (sd *SubscriptionDispatcher) Update(id string, sr SubscriptionReq) (SubscriptionReq, bool) {
	if s, found := sd.subscriptions.Get(id); found {
		mdclog(MdclogInfo, fmt.Sprintf("Subscription id=%s found: %v ... updating", id, s.(Subscription).req))

		sr.Id = id
		sd.subscriptions.Set(id, Subscription{sr, s.(Subscription).resp})
		sd.db.StoreSubscriptions(sd.subscriptions)

		return sr, found
	}
	return SubscriptionReq{}, false
}

func (sd *SubscriptionDispatcher) Publish(x Xapp, et EventType) {
	sd.notifyClients([]Xapp{x}, et)
}

func (sd *SubscriptionDispatcher) notifyClients(xapps []Xapp, et EventType) {
	if len(xapps) == 0 || len(sd.subscriptions) == 0 {
		mdclog(MdclogInfo, fmt.Sprintf("Nothing to publish [%d:%d]", len(xapps), len(sd.subscriptions)))
		return
	}

	sd.Seq = sd.Seq + 1
	for v := range sd.subscriptions.Iter() {
		go sd.notify(xapps, et, v.Val.(Subscription), sd.Seq)
	}
}

func (sd *SubscriptionDispatcher) notify(xapps []Xapp, et EventType, s Subscription, seq int) error {
	notif := []SubscriptionNotif{}
	notif = append(notif, SubscriptionNotif{Id: s.req.Id, Version: seq, EventType: string(et), XappData: xapps})

	jsonData, err := json.Marshal(notif)
	if err != nil {
		mdclog(MdclogInfo, fmt.Sprintf("json.Marshal failed: %v", err))
		return err
	}

	// Execute the request with retry policy
	return sd.retry(s, func() error {
		resp, err := http.Post(s.req.TargetUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			mdclog(MdclogInfo, fmt.Sprintf("Posting to subscription failed: %v", err))
			return err
		}

		if resp.StatusCode != http.StatusOK {
			mdclog(MdclogInfo, fmt.Sprintf("Client returned error code: %d", resp.StatusCode))
			return err
		}

		mdclog(MdclogInfo, fmt.Sprintf("subscription to '%s' dispatched, response code: %d \n", s.req.TargetUrl, resp.StatusCode))
		return nil
	})
}

func (sd *SubscriptionDispatcher) retry(s Subscription, fn func() error) error {
	if err := fn(); err != nil {
		// Todo: use exponential backoff, or similar mechanism
		if s.req.MaxRetries--; s.req.MaxRetries > 0 {
			time.Sleep(time.Duration(s.req.RetryTimer) * time.Second)
			return sd.retry(s, fn)
		}
		sd.subscriptions.Remove(s.req.Id)
		return err
	}
	return nil
}
