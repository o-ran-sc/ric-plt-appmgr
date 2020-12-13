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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"strconv"
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
)

var rh *Resthook
var resp models.SubscriptionResponse
var mockedSdl *SdlMock
var mockedSdl2 *SdlMock

// Test cases
func TestMain(m *testing.M) {
	appmgr.Init()
	appmgr.Logger.SetLevel(0)

	mockedSdl = new(SdlMock)
	mockedSdl2 = new(SdlMock)
	rh = createResthook(false, mockedSdl,mockedSdl2)
	code := m.Run()
	os.Exit(code)
}

func TestAddSubscriptionSuccess(t *testing.T) {
	var mockSdlRetOk error
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")

	mockedSdl.expectDbSet(t, subsReq, mockSdlRetOk)
	resp := rh.AddSubscription(subsReq)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeCreated)
}

func TestAddSubscriptionExists(t *testing.T) {
	resp := rh.AddSubscription(createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeCreated)
}

func TestDeletesubscriptionSuccess(t *testing.T) {
	var mockSdlRetOk error

	mockedSdl.On("Set", mock.Anything).Return(mockSdlRetOk)
	resp := rh.AddSubscription(createSubscription(models.EventTypeDeleted, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeDeleted)

	resp, ok := rh.DeleteSubscription(resp.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeDeleted)
}

func TestDeletesubscriptionInvalid(t *testing.T) {
	resp, ok := rh.DeleteSubscription("Non-existent-ID")
	assert.Equal(t, ok, false)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventType(""))
}

func TestModifySubscriptionSuccess(t *testing.T) {
	resp := rh.AddSubscription(createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeCreated)

	resp, ok := rh.ModifySubscription(resp.ID, createSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, ok, true)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventTypeModified)
}

func TestModifySubscriptionForNonExistingSubscription(t *testing.T) {
	resp, ok := rh.ModifySubscription("Non-existent-ID", createSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))
	assert.Equal(t, ok, false)
	assert.Equal(t, *resp, models.SubscriptionResponse{})
}

func TestDeleteSubscriptionForNonExistingSubscription(t *testing.T) {
	resp, ok := rh.DeleteSubscription("Non-existent-ID")
	assert.Equal(t, ok, false)
	assert.Equal(t, resp.Version, int64(0))
	assert.Equal(t, resp.EventType, models.EventType(""))
}

func TestGetAllSubscriptionSuccess(t *testing.T) {
	flushExistingSubscriptions()
	subscriptions := rh.GetAllSubscriptions()
	assert.Equal(t, len(subscriptions), 0)

	rh.AddSubscription(createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook"))
	rh.AddSubscription(createSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2"))

	subscriptions = rh.GetAllSubscriptions()
	assert.Equal(t, len(subscriptions), 2)
}

func TestGetSubscriptionByIdSuccess(t *testing.T) {
	flushExistingSubscriptions()

	sub1 := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	sub2 := createSubscription(models.EventTypeModified, int64(5), int64(10), "http://localhost:8087/xapps_hook2")
	r1 := rh.AddSubscription(sub1)
	r2 := rh.AddSubscription(sub2)

	resp1, ok := rh.GetSubscriptionById(r1.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp1.Data, sub1.Data)

	resp2, ok := rh.GetSubscriptionById(r2.ID)
	assert.Equal(t, ok, true)
	assert.Equal(t, resp2.Data, sub2.Data)
}

func TestGetSubscriptionByIdForNonExistingSubscription(t *testing.T) {
	resp, ok := rh.GetSubscriptionById("Non-existent-ID")
	assert.Equal(t, ok, false)
	assert.Equal(t, resp, models.Subscription{})
}

func TestNotifyClientsNoXapp(t *testing.T) {
	rh.NotifyClients(models.AllDeployedXapps{}, models.EventTypeUndeployed)
}

func TestNotifySuccess(t *testing.T) {
	flushExistingSubscriptions()

	sub := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	resp := rh.AddSubscription(sub)

	xapp := getDummyXapp()
	ts := createHTTPServer(t, "POST", "/xapps_hook", 8087, http.StatusOK, nil)
	defer ts.Close()

	v, ok := rh.subscriptions.Get(resp.ID)
	assert.True(t, ok)
	err := rh.notify(models.AllDeployedXapps{&xapp}, models.EventTypeUndeployed, v.(SubscriptionInfo), 1)
	assert.Nil(t, err)
}

func TestNotifySuccessIfHttpErrorResponse(t *testing.T) {
	flushExistingSubscriptions()

	sub := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	resp := rh.AddSubscription(sub)

	xapp := getDummyXapp()
	ts := createHTTPServer(t, "POST", "/xapps_hook", 8087, http.StatusInternalServerError, nil)
	defer ts.Close()

	v, ok := rh.subscriptions.Get(resp.ID)
	assert.True(t, ok)
	err := rh.notify(models.AllDeployedXapps{&xapp}, models.EventTypeUndeployed, v.(SubscriptionInfo), 1)
	assert.Nil(t, err)
}

func TestNotifyReturnsErrorAfterRetriesIfNoHttpServer(t *testing.T) {
	flushExistingSubscriptions()

	sub := createSubscription(models.EventTypeCreated, int64(2), int64(1), "http://localhost:8087/xapps_hook")
	resp := rh.AddSubscription(sub)

	xapp := getDummyXapp()

	v, ok := rh.subscriptions.Get(resp.ID)
	assert.True(t, ok)
	err := rh.notify(models.AllDeployedXapps{&xapp}, models.EventTypeUndeployed, v.(SubscriptionInfo), 1)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(rh.subscriptions.Items()))
}

func TestRestoreSubscriptionsSuccess(t *testing.T) {
	var mockSdlRetOk error
	mSdl := new(SdlMock)
	mSdl2 := new(SdlMock)
	key := "key-1"

	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	mockSdlGetRetVal := make(map[string]interface{})
	//Cast data to string to act like a real SDL/Redis client
	mockSdlGetRetVal[key] = string(serializedSubsReq)
	mSdl.On("GetAll").Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", []string{key}).Return(mockSdlGetRetVal, mockSdlRetOk).Once()
	restHook := createResthook(true, mSdl,mSdl2)

	val, found := restHook.subscriptions.Get(key)
	assert.True(t, found)
	assert.Equal(t, subsReq, val.(SubscriptionInfo).req)
}

func TestRestoreSubscriptionsFailsIfSdlGetAllFails(t *testing.T) {
	var mockSdlRetStatus error
	mSdl := new(SdlMock)
	mSdl2 := new(SdlMock)
	getCalled := 0
	mGetAllCall := mSdl.On("GetAll")
	mGetAllCall.RunFn = func(args mock.Arguments) {
		if getCalled > 0 {
			mockSdlRetStatus = errors.New("some SDL error")
		}
		getCalled++
		mGetAllCall.ReturnArguments = mock.Arguments{[]string{}, mockSdlRetStatus}
	}

	restHook := createResthook(true, mSdl,mSdl2)
	assert.Equal(t, 0, len(restHook.subscriptions.Items()))
}

func TestRestoreSubscriptionsFailsIfSdlGetFails(t *testing.T) {
	var mockSdlRetOk error
	mSdl := new(SdlMock)
	mSdl2 := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")
	key := "key-1"
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mSdl.On("GetAll").Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()

	restHook := createResthook(true, mSdl,mSdl2)
	assert.Equal(t, 0, len(restHook.subscriptions.Items()))
}

func TestTeardown(t *testing.T) {
	var mockSdlRetOk error
	mockedSdl.On("RemoveAll").Return(mockSdlRetOk).Once()

	rh.FlushSubscriptions()
}

func createSubscription(et models.EventType, maxRetries, retryTimer int64, targetUrl string) models.SubscriptionRequest {
	return models.SubscriptionRequest{&models.SubscriptionData{et, &maxRetries, &retryTimer, &targetUrl}}
}

func getDummyXapp() models.Xapp {
	return generateXapp("dummy-xapp", "deployed", "1.0", "dummy-xapp-8984fc9fd-bkcbp", "running", "service-ricxapp-dummy-xapp-rmr.ricxapp", "4560")
}

func generateXapp(name, status, ver, iname, istatus, ip, port string) (x models.Xapp) {
	x.Name = &name
	x.Status = status
	x.Version = ver
	p, _ := strconv.Atoi(port)
	var msgs appmgr.RtmData

	instance := &models.XappInstance{
		Name:       &iname,
		Status:     istatus,
		IP:         ip,
		Port:       int64(p),
		TxMessages: msgs.TxMessages,
		RxMessages: msgs.RxMessages,
	}
	x.Instances = append(x.Instances, instance)
	return
}

func flushExistingSubscriptions() {
	var mockSdlRetOk error
	mockedSdl.On("RemoveAll").Return(mockSdlRetOk).Once()
	rh.FlushSubscriptions()
}

func createHTTPServer(t *testing.T, method, url string, port, status int, respData interface{}) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Error("Failed to create listener: " + err.Error())
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, method)
		assert.Equal(t, r.URL.String(), url)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		b, _ := json.Marshal(respData)
		w.Write(b)
	}))
	ts.Listener.Close()
	ts.Listener = l

	ts.Start()
	time.Sleep(time.Duration(1 * time.Second))
	return ts
}

func (m *SdlMock) expectDbSet(t *testing.T, subsReq models.SubscriptionRequest, mockRet error) {
	serializedSubReq, _ := json.Marshal(subsReq)
	m.On("Set", mock.Anything).Run(
		func(args mock.Arguments) {
			sdlKVs := args.Get(0).([]interface{})
			assert.Equal(t, 2, len(sdlKVs))
			//Validate that subscription request is set to SDL
			assert.Equal(t, serializedSubReq, sdlKVs[1])
		}).Return(mockRet).Once()
}

type SdlMock struct {
	mock.Mock
}

func (m *SdlMock) Set(pairs ...interface{}) error {
	a := m.Called(pairs)
	return a.Error(0)
}

func (m *SdlMock) Get(keys []string) (map[string]interface{}, error) {
	a := m.Called(keys)
	return a.Get(0).(map[string]interface{}), a.Error(1)
}

func (m *SdlMock) GetAll() ([]string, error) {
	a := m.Called()
	return a.Get(0).([]string), a.Error(1)
}

func (m *SdlMock) RemoveAll() error {
	a := m.Called()
	return a.Error(0)
}

func (m *SdlMock) Remove(keys []string) error {
	a := m.Called()
	return a.Error(0)
}
