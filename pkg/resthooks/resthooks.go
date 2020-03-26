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
	"bytes"
	"encoding/json"
	sdl "gerrit.o-ran-sc.org/r/ric-plt/sdlgo"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/segmentio/ksuid"
	"net/http"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
)

func NewResthook(restoreData bool) *Resthook {
	rh := &Resthook{
		client: &http.Client{},
		db:     sdl.NewSdlInstance("appmgr", sdl.NewDatabase()),
	}

	if restoreData {
		rh.subscriptions = rh.RestoreSubscriptions()
	} else {
		rh.subscriptions = cmap.New()
	}
	return rh
}

func (rh *Resthook) AddSubscription(sr models.SubscriptionRequest) *models.SubscriptionResponse {
	for v := range rh.subscriptions.IterBuffered() {
		r := v.Val.(SubscriptionInfo).req
		if *r.Data.TargetURL == *sr.Data.TargetURL && r.Data.EventType == sr.Data.EventType {
			appmgr.Logger.Info("Similar subscription already exists!")
			resp := v.Val.(SubscriptionInfo).resp
			return &resp
		}
	}

	key := ksuid.New().String()
	resp := models.SubscriptionResponse{ID: key, Version: 0, EventType: sr.Data.EventType}
	rh.subscriptions.Set(key, SubscriptionInfo{key, sr, resp})
	rh.StoreSubscriptions(rh.subscriptions)

	appmgr.Logger.Info("Sub: New subscription added: key=%s targetUl=%s eventType=%s", key, *sr.Data.TargetURL, sr.Data.EventType)
	return &resp
}

func (rh *Resthook) DeleteSubscription(id string) (*models.SubscriptionResponse, bool) {
	if v, found := rh.subscriptions.Get(id); found {
		appmgr.Logger.Info("Subscription id=%s found: %v ... deleting", id, v.(SubscriptionInfo).req)

		rh.subscriptions.Remove(id)
		rh.StoreSubscriptions(rh.subscriptions)
		resp := v.(SubscriptionInfo).resp
		return &resp, found
	}
	return &models.SubscriptionResponse{}, false
}

func (rh *Resthook) ModifySubscription(id string, req models.SubscriptionRequest) (*models.SubscriptionResponse, bool) {
	if s, found := rh.subscriptions.Get(id); found {
		appmgr.Logger.Info("Subscription id=%s found: %v ... updating", id, s.(SubscriptionInfo).req)

		resp := models.SubscriptionResponse{ID: id, Version: 0, EventType: req.Data.EventType}
		rh.subscriptions.Set(id, SubscriptionInfo{id, req, resp})
		rh.StoreSubscriptions(rh.subscriptions)

		return &resp, found
	}
	return &models.SubscriptionResponse{}, false
}

func (rh *Resthook) GetAllSubscriptions() (hooks models.AllSubscriptions) {
	hooks = models.AllSubscriptions{}
	for v := range rh.subscriptions.IterBuffered() {
		s := v.Val.(SubscriptionInfo)
		r := v.Val.(SubscriptionInfo).req
		hooks = append(hooks, &models.Subscription{&models.SubscriptionData{r.Data.EventType, r.Data.MaxRetries, r.Data.RetryTimer, r.Data.TargetURL}, s.Id})
	}

	return hooks
}

func (rh *Resthook) GetSubscriptionById(id string) (models.Subscription, bool) {
	if v, found := rh.subscriptions.Get(id); found {
		appmgr.Logger.Info("Subscription id=%s found: %v", id, v.(SubscriptionInfo).req)
		r := v.(SubscriptionInfo).req
		return models.Subscription{&models.SubscriptionData{r.Data.EventType, r.Data.MaxRetries, r.Data.RetryTimer, r.Data.TargetURL}, id}, found
	}
	return models.Subscription{}, false
}

func (rh *Resthook) PublishSubscription(x models.Xapp, et models.EventType) {
	rh.NotifyClients(models.AllDeployedXapps{&x}, et)
}

func (rh *Resthook) NotifyClients(xapps models.AllDeployedXapps, et models.EventType) {
	if len(xapps) == 0 || len(rh.subscriptions) == 0 {
		appmgr.Logger.Info("Nothing to publish [%d:%d]", len(xapps), len(rh.subscriptions))
		return
	}

	rh.Seq = rh.Seq + 1
	for v := range rh.subscriptions.Iter() {
		go rh.notify(xapps, et, v.Val.(SubscriptionInfo), rh.Seq)
	}
}

func (rh *Resthook) notify(xapps models.AllDeployedXapps, et models.EventType, s SubscriptionInfo, seq int64) error {
	xappData, err := json.Marshal(xapps)
	if err != nil {
		appmgr.Logger.Info("json.Marshal failed: %v", err)
		return err
	}

	// TODO: Use models.SubscriptionNotification instead of internal ...
	notif := SubscriptionNotification{ID: s.Id, Version: seq, Event: string(et), XApps: string(xappData)}
	jsonData, err := json.Marshal(notif)
	if err != nil {
		appmgr.Logger.Info("json.Marshal failed: %v", err)
		return err
	}

	// Execute the request with retry policy
	return rh.retry(s, func() error {
		appmgr.Logger.Info("Posting notification to TargetURL=%s: %v", *s.req.Data.TargetURL, notif)
		resp, err := http.Post(*s.req.Data.TargetURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			appmgr.Logger.Info("Posting to subscription failed: %v", err)
			return err
		}

		if resp.StatusCode != http.StatusOK {
			appmgr.Logger.Info("Client returned error code: %d", resp.StatusCode)
			return err
		}

		appmgr.Logger.Info("subscription to '%s' dispatched, response code: %d", *s.req.Data.TargetURL, resp.StatusCode)
		return nil
	})
}

func (rh *Resthook) retry(s SubscriptionInfo, fn func() error) error {
	if err := fn(); err != nil {
		// Todo: use exponential backoff, or similar mechanism
		if *s.req.Data.MaxRetries--; *s.req.Data.MaxRetries > 0 {
			time.Sleep(time.Duration(*s.req.Data.RetryTimer) * time.Second)
			return rh.retry(s, fn)
		}
		rh.subscriptions.Remove(s.Id)
		return err
	}
	return nil
}

func (rh *Resthook) StoreSubscriptions(m cmap.ConcurrentMap) {
	for v := range m.Iter() {
		s := v.Val.(SubscriptionInfo)

		data, err := json.Marshal(s.req)
		if err != nil {
			appmgr.Logger.Error("json.marshal failed: %v ", err.Error())
			return
		}

		if err := rh.db.Set(s.Id, data); err != nil {
			appmgr.Logger.Error("DB.session.Set failed: %v ", err.Error())
		}
	}
}

func (rh *Resthook) RestoreSubscriptions() (m cmap.ConcurrentMap) {
	rh.VerifyDBConnection()

	m = cmap.New()
	keys, err := rh.db.GetAll()
	if err != nil {
		appmgr.Logger.Error("DB.session.GetAll failed: %v ", err.Error())
		return
	}

	for _, key := range keys {
		value, err := rh.db.Get([]string{key})
		if err != nil {
			appmgr.Logger.Error("DB.session.Get failed: %v ", err.Error())
			return
		}

		var item models.SubscriptionRequest
		if err = json.Unmarshal([]byte(value[key].(string)), &item); err != nil {
			appmgr.Logger.Error("json.Unmarshal failed: %v ", err.Error())
			return
		}

		resp := models.SubscriptionResponse{ID: key, Version: 0, EventType: item.Data.EventType}
		m.Set(key, SubscriptionInfo{key, item, resp})
	}

	return m
}

func (rh *Resthook) VerifyDBConnection() {
	// Test DB connection, and wait until ready!
	for {
		if _, err := rh.db.GetAll(); err == nil {
			return
		}
		appmgr.Logger.Error("Database connection not ready, waiting ...")
		time.Sleep(time.Duration(5 * time.Second))
	}
}

func (rh *Resthook) FlushSubscriptions() {
	rh.db.RemoveAll()
	rh.subscriptions = cmap.New()
}
