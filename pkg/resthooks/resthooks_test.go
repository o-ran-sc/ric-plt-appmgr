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
	"strconv"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
)

var rh *Resthook
var resp models.SubscriptionResponse
var mockedSdl *SdlMock

// Test cases
func TestMain(m *testing.M) {
	appmgr.Init()
	appmgr.Logger.SetLevel(0)

	mockedSdl = new(SdlMock)
	NewResthook(false)
	rh = createResthook(false, mockedSdl)
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

	mockedSdl.On("Set", appmgrSdlNs, mock.Anything).Return(mockSdlRetOk)
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
	key := "key-1"

	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	mockSdlGetRetVal := make(map[string]interface{})
	//Cast data to string to act like a real SDL/Redis client
	mockSdlGetRetVal[key] = string(serializedSubsReq)
	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetOk).Once()
	restHook := createResthook(true, mSdl)

	val, found := restHook.subscriptions.Get(key)
	assert.True(t, found)
	assert.Equal(t, subsReq, val.(SubscriptionInfo).req)
}

func TestRestoreSubscriptionsFailsIfSdlGetAllFails(t *testing.T) {
	var mockSdlRetStatus error
	mSdl := new(SdlMock)
	getCalled := 0
	mGetAllCall := mSdl.On("GetAll", appmgrSdlNs)
	mGetAllCall.RunFn = func(args mock.Arguments) {
		if getCalled > 0 {
			mockSdlRetStatus = errors.New("some SDL error")
		}
		getCalled++
		mGetAllCall.ReturnArguments = mock.Arguments{[]string{}, mockSdlRetStatus}
	}

	restHook := createResthook(true, mSdl)
	assert.Equal(t, 0, len(restHook.subscriptions.Items()))
}

func TestRestoreSubscriptionsFailsIfSdlGetFails(t *testing.T) {
	var mockSdlRetOk error
	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")
	key := "key-1"
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()

	restHook := createResthook(true, mSdl)
	assert.Equal(t, 0, len(restHook.subscriptions.Items()))
}

func TestTeardown(t *testing.T) {
	var mockSdlRetOk error
	mockedSdl.On("RemoveAll", appmgrSdlNs).Return(mockSdlRetOk).Once()

	rh.FlushSubscriptions()
}

func TestUpdateAppDataFail1(t *testing.T) {
	var reg models.RegisterRequest
	var tEndpoint string = "10.104.237.59"
	reg.HTTPEndpoint = &tEndpoint
	rh.UpdateAppData(reg, false)
}

func TestUpdateAppDataFail2(t *testing.T) {
	var mockSdlRetOk error
	var params models.RegisterRequest

	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")
	var tEndpoint1 string = "10.104.237.59:8087"
	params.HTTPEndpoint = &tEndpoint1
	serializedSubsReq2, err := json.Marshal(params)
	if err != nil {
		t.Logf("error in marshal .. %v", err)
	}

	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	key := "key-1"
	value := "endpoints"
	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mockSdlGetRetVal2 := make(map[string]interface{})
	mockSdlGetRetVal2[value] = serializedSubsReq2
	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()
	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal2, mockSdlRetOk).Once()

	restHook := createResthook(false, mSdl)

	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal2, mockSdlRetOk).Once()

	ret := restHook.GetAppsInSDL()
	if ret == nil {
		assert.Nil(t, ret)
	}
}

func TestGetAppsInSDLFail3(t *testing.T) {
	var mockSdlRetOk error
	var params models.RegisterRequest

	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")

	serializedSubsReq1, err := json.Marshal(params)
	if err != nil {
		t.Logf("error in marshal .. %v", err)
	}

	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	key := "key-1"
	value := "endpoints"
	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mockSdlGetRetVal1 := make(map[string]interface{})
	mockSdlGetRetVal1[key] = serializedSubsReq1

	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()
	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()

	restHook := createResthook(false, mSdl)

	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()
	ret2 := restHook.GetAppsInSDL()
	if ret2 != nil {
		t.Logf("SDL Returning: %s \n", *ret2)
	} else {
		assert.Nil(t, ret2)
	}
}

func TestUpdateAppDataSucc(t *testing.T) {
	var mockSdlRetOk error
	var params models.RegisterRequest

	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")

	var tEndpoint1 string = "10.104.237.59:8087"
	params.HTTPEndpoint = &tEndpoint1
	serializedSubsReq1, err := json.Marshal(params)
	if err != nil {
		t.Logf("error in marshal .. %v", err)
	}
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	key := "key-1"
	value := "endpoints"
	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mockSdlGetRetVal1 := make(map[string]interface{})
	mockSdlGetRetVal1[key] = serializedSubsReq1

	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()
	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()

	restHook := createResthook(false, mSdl)

	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()
	mSdl.On("Set", appDbSdlNs, mock.Anything).Return(mockSdlRetOk)
	restHook.UpdateAppData(params, true)
}

func TestUpdateAppDataSucc1(t *testing.T) {
	var mockSdlRetOk error
	var params models.RegisterRequest

	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")

	var tEndpoint1 string = "10.104.237.59:8087"
	params.HTTPEndpoint = &tEndpoint1
	appsindb := []string{"10.104.237.59:8088 ", " ", " ", " 10.104.237.59:8087"}
	serializedSubsReq1, err := json.Marshal(appsindb)
	if err != nil {
		t.Logf("error in marshal .. %v", err)
	}
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	key := "key-1"
	value := "endpoints"
	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mockSdlGetRetVal1 := make(map[string]interface{})
	mockSdlGetRetVal1[value] = serializedSubsReq1

	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()
	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()

	restHook := createResthook(false, mSdl)

	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()
	mSdl.On("Set", appDbSdlNs, []string{value}).Return(mockSdlRetOk).Twice()

	mSdl.On("Remove", appDbSdlNs, mock.Anything).Return(mockSdlRetOk)
	mSdl.On("Set", appDbSdlNs, mock.Anything).Return(mockSdlRetOk)
	restHook.UpdateAppData(params, true)
}

func TestUpdateAppDataSucc2(t *testing.T) {
	var mockSdlRetOk error
	var params models.RegisterRequest

	mSdl := new(SdlMock)
	mockSdlRetNok := errors.New("some SDL error")

	var tEndpoint1 string = "10.104.237.59:8087"
	params.Config = "/temp/config.yaml"
	params.HTTPEndpoint = &tEndpoint1
	serializedSubsReq1, err := json.Marshal(tEndpoint1)
	if err != nil {
		t.Logf("error in marshal .. %v", err)
	}
	subsReq := createSubscription(models.EventTypeCreated, int64(5), int64(10), "http://localhost:8087/xapps_hook")
	serializedSubsReq, err := json.Marshal(subsReq)
	assert.Nil(t, err)

	key := "key-1"
	value := "endpoints"
	mockSdlGetRetVal := make(map[string]interface{})
	mockSdlGetRetVal[key] = serializedSubsReq

	mockSdlGetRetVal1 := make(map[string]interface{})
	mockSdlGetRetVal1[value] = serializedSubsReq1

	mSdl.On("GetAll", appmgrSdlNs).Return([]string{key}, mockSdlRetOk).Twice()
	mSdl.On("Get", appmgrSdlNs, []string{key}).Return(mockSdlGetRetVal, mockSdlRetNok).Once()
	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()

	restHook := createResthook(false, mSdl)

	mSdl.On("Get", appDbSdlNs, []string{value}).Return(mockSdlGetRetVal1, mockSdlRetOk).Once()
	mSdl.On("Set", appDbSdlNs, []string{value}).Return(mockSdlRetOk).Twice()

	mSdl.On("Remove", appDbSdlNs, mock.Anything).Return(mockSdlRetOk)
	mSdl.On("Set", appDbSdlNs, mock.Anything).Return(mockSdlRetOk)
	restHook.UpdateAppData(params, true)
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
	mockedSdl.On("RemoveAll", appmgrSdlNs).Return(mockSdlRetOk).Once()
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
	m.On("Set", appmgrSdlNs, mock.Anything).Run(
		func(args mock.Arguments) {
			sdlKVs := args.Get(1).([]interface{})
			assert.Equal(t, 2, len(sdlKVs))
			//Validate that subscription request is set to SDL
			assert.Equal(t, serializedSubReq, sdlKVs[1])
		}).Return(mockRet).Once()
}

func TestPublishSubscription(t *testing.T) {
	sub := createSubscription(models.EventTypeCreated, int64(2), int64(1), "http://localhost:8087/xapps_hook")
	resp := rh.AddSubscription(sub)

	xapp := getDummyXapp()

	v, ok := rh.subscriptions.Get(resp.ID)
	assert.True(t, ok)
	if v == nil {
		t.Logf("value : %+v", v)
	}
	rh.PublishSubscription(xapp, models.EventTypeUndeployed)
}

type SdlMock struct {
	mock.Mock
}

func (m *SdlMock) Set(ns string, pairs ...interface{}) error {
	a := m.Called(ns, pairs)
	return a.Error(0)
}

func (m *SdlMock) Get(ns string, keys []string) (map[string]interface{}, error) {
	a := m.Called(ns, keys)
	return a.Get(0).(map[string]interface{}), a.Error(1)
}

func (m *SdlMock) GetAll(ns string) ([]string, error) {
	a := m.Called(ns)
	return a.Get(0).([]string), a.Error(1)
}

func (m *SdlMock) RemoveAll(ns string) error {
	a := m.Called(ns)
	return a.Error(0)
}

func (m *SdlMock) Remove(ns string, keys []string) error {
	a := m.Called(ns)
	return a.Error(0)
}
