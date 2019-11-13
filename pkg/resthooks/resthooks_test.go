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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
)

var resp models.SubscriptionResponse

// Test cases
func TestMain(m *testing.M) {
	appmgr.Init()
	appmgr.Logger.SetLevel(0)

	code := m.Run()
	os.Exit(code)
}

func TestAddSubscriptionSuccess(t *testing.T) {
	resp := NewResthook().AddSubscription(CreateSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeCreated)
}

func TestAddSubscriptionExists(t *testing.T) {
	resp := NewResthook().AddSubscription(CreateSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventType(""))
}

func TestDeletesubscriptionSuccess(t *testing.T) {
	resp := NewResthook().AddSubscription(CreateSubscription(models.EventTypeDeleted, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeDeleted)

	resp, ok := NewResthook().DeleteSubscription(resp.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeDeleted)
}

func TestDeletesubscriptionInvalid(t *testing.T) {
	resp, ok := NewResthook().DeleteSubscription("Non-existent-ID")
	assert.Equal(t, ok, false)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventType(""))
}

func TestModifySubscriptionSuccess(t *testing.T) {
	resp := NewResthook().AddSubscription(CreateSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeCreated)

	resp, ok := NewResthook().ModifySubscription(resp.ID, CreateSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, ok, true)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeModified)
}

func TestModifysubscriptionInvalid(t *testing.T) {
	resp, ok := NewResthook().DeleteSubscription("Non-existent-ID")
	assert.Equal(t, ok, false)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventType(""))
}

func TestGetAllSubscriptionSuccess(t *testing.T) {
	NewResthook().FlushSubscriptions()
	subscriptions := NewResthook().GetAllSubscriptions()
	assert.Equal(t, len(subscriptions), 0)

	NewResthook().AddSubscription(CreateSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook"))
	NewResthook().AddSubscription(CreateSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))

	subscriptions = NewResthook().GetAllSubscriptions()
	assert.Equal(t, len(subscriptions), 2)
}

func TestGetSubscriptionByIdSuccess(t *testing.T) {
	NewResthook().FlushSubscriptions()
	sub1 := CreateSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	sub2 := CreateSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2")
	r1 := NewResthook().AddSubscription(sub1)
	r2 := NewResthook().AddSubscription(sub2)

	resp1, ok := NewResthook().GetSubscriptionById(r1.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp1.Data, sub1.Data)

	resp2, ok := NewResthook().GetSubscriptionById(r2.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp2.Data, sub2.Data)
}

func TestTeardown(t *testing.T) {
	NewResthook().FlushSubscriptions()
}

func CreateSubscription(et models.EventType, maxRetries, retryTimer int64, targetUrl string) models.SubscriptionRequest {
	return models.SubscriptionRequest{&models.SubscriptionData{et, &maxRetries, &retryTimer, &targetUrl}}
}
