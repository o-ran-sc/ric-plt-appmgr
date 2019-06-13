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
	"encoding/json"
	sdl "gerrit.oran-osc.org/r/ric-plt/sdlgo"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/spf13/viper"
	"time"
)

type DB struct {
	session *sdl.SdlInstance
}

func (d *DB) Create() {
	ns := viper.GetString("db.sessionNamespace")
	d.session = sdl.NewSdlInstance(ns, sdl.NewDatabase())

	// Test DB connection, and wait until ready!
	for {
		if _, err := d.session.GetAll(); err == nil {
			return
		}
		Logger.Error("Database connection not ready, waiting ...")
		time.Sleep(time.Duration(5 * time.Second))
	}
}

func (d *DB) StoreSubscriptions(m cmap.ConcurrentMap) {
	for v := range m.Iter() {
		s := v.Val.(Subscription)

		data, err := json.Marshal(s.req)
		if err != nil {
			Logger.Error("json.marshal failed: %v ", err.Error())
			return
		}

		if err := d.session.Set(s.req.Id, data); err != nil {
			Logger.Error("DB.session.Set failed: %v ", err.Error())
		}
	}
}

func (d *DB) RestoreSubscriptions() (m cmap.ConcurrentMap) {
	m = cmap.New()

	keys, err := d.session.GetAll()
	if err != nil {
		Logger.Error("DB.session.GetAll failed: %v ", err.Error())
		return
	}

	for _, key := range keys {
		value, err := d.session.Get([]string{key})
		if err != nil {
			Logger.Error("DB.session.Get failed: %v ", err.Error())
			return
		}

		var item SubscriptionReq
		if err = json.Unmarshal([]byte(value[key].(string)), &item); err != nil {
			Logger.Error("json.Unmarshal failed: %v ", err.Error())
			return
		}

		resp := SubscriptionResp{key, 0, item.EventType}
		m.Set(key, Subscription{item, resp})
	}

	return m
}
