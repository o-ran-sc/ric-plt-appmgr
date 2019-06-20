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
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var x XappManager
var xapp Xapp
var xapps []Xapp
var helmError error

type MockedHelmer struct {
}

func (h *MockedHelmer) SetCM(cm ConfigMapper) {
}

func (sd *MockedHelmer) Initialize() {
}

func (h *MockedHelmer) Status(name string) (Xapp, error) {
	return xapp, helmError
}

func (h *MockedHelmer) StatusAll() ([]Xapp, error) {
	return xapps, helmError
}

func (h *MockedHelmer) SearchAll() (s []string) {
	return s
}

func (h *MockedHelmer) List() (names []string, err error) {
	return names, helmError
}

func (h *MockedHelmer) Install(m XappDeploy) (Xapp, error) {
	return xapp, helmError
}

func (h *MockedHelmer) Delete(name string) (Xapp, error) {
	return xapp, helmError
}

// Test cases
func TestMain(m *testing.M) {
	Logger = NewLogger("xapp-manager")
	loadConfig()

	xapp = Xapp{}
	xapps = []Xapp{}

	cm := MockedConfigMapper{}
	h := MockedHelmer{}
	x = XappManager{}
	x.Initialize(&h, &cm)

	// Just run on the background (for coverage)
	go x.Run()
	x.ready = true

	time.Sleep(time.Duration(2 * time.Second))

	code := m.Run()
	os.Exit(code)
}

func TestGetHealthCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ric/v1/health/ready", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetAppsReturnsEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ric/v1/xapps", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	if body := response.Body.String(); body != "[]" {
		t.Errorf("handler returned unexpected body: got %v want []", body)
	}
}

func TestCreateXApp(t *testing.T) {
	xapp = generateXapp("dummy-xapp", "started", "1.0", "dummy-xapp-1234-5678", "running", "127.0.0.1", "9999")

	payload := []byte(`{"name":"dummy-xapp"}`)
	req, _ := http.NewRequest("POST", "/ric/v1/xapps", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseData(t, response, http.StatusCreated, false)
}

func TestGetAppsReturnsListOfXapps(t *testing.T) {
	xapps = append(xapps, xapp)
	req, _ := http.NewRequest("GET", "/ric/v1/xapps", nil)
	response := executeRequest(req)

	checkResponseData(t, response, http.StatusOK, true)
}

func TestGetAppByIdReturnsGivenXapp(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ric/v1/xapps/"+xapp.Name, nil)
	response := executeRequest(req)

	checkResponseData(t, response, http.StatusOK, false)
}

func TestGetAppInstanceByIdReturnsGivenXapp(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ric/v1/xapps/"+xapp.Name+"/instances/dummy-xapp-1234-5678", nil)
	response := executeRequest(req)

	var ins XappInstance
	checkResponseCode(t, http.StatusOK, response.Code)
	json.NewDecoder(response.Body).Decode(&ins)

	if !reflect.DeepEqual(ins, xapp.Instances[0]) {
		t.Errorf("handler returned unexpected body: got: %v, expected: %v", ins, xapp.Instances[0])
	}
}

func TestDeleteAppRemovesGivenXapp(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/ric/v1/xapps/"+xapp.Name, nil)
	response := executeRequest(req)

	checkResponseData(t, response, http.StatusNoContent, false)

	// Xapp not found from the Redis DB
	helmError = errors.New("Not found")

	req, _ = http.NewRequest("GET", "/ric/v1/xapps/"+xapp.Name, nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetConfigReturnsEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ric/v1/config", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateConfigFailsWithMethodNotAllowed(t *testing.T) {
	req, _ := http.NewRequest("POST", "/ric/v1/config", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

func TestCreateConfigOk(t *testing.T) {
	payload := []byte(`{"name":"dummy-xapp"}`)
	req, _ := http.NewRequest("POST", "/ric/v1/config", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)
}

func TestDeleteConfigOk(t *testing.T) {
	payload := []byte(`{"name":"dummy-xapp"}`)
	req, _ := http.NewRequest("DELETE", "/ric/v1/config", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNoContent, response.Code)
}

// Error handling
func TestGetXappReturnsError(t *testing.T) {
	helmError = errors.New("Not found")

	req, _ := http.NewRequest("GET", "/ric/v1/xapps/invalidXappName", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetXappInstanceReturnsError(t *testing.T) {
	helmError = errors.New("Some error")

	req, _ := http.NewRequest("GET", "/ric/v1/xapps/"+xapp.Name+"/instances/invalidXappName", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetXappListReturnsError(t *testing.T) {
	helmError = errors.New("Internal error")

	req, _ := http.NewRequest("GET", "/ric/v1/xapps", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusInternalServerError, response.Code)
}

func TestCreateXAppWithoutXappData(t *testing.T) {
	req, _ := http.NewRequest("POST", "/ric/v1/xapps", nil)
	response := executeRequest(req)
	checkResponseData(t, response, http.StatusMethodNotAllowed, false)
}

func TestCreateXAppWithInvalidXappData(t *testing.T) {
	body := []byte("Invalid JSON data ...")

	req, _ := http.NewRequest("POST", "/ric/v1/xapps", bytes.NewBuffer(body))
	response := executeRequest(req)
	checkResponseData(t, response, http.StatusMethodNotAllowed, false)
}

func TestCreateXAppReturnsError(t *testing.T) {
	helmError = errors.New("Not found")

	payload := []byte(`{"name":"dummy-xapp"}`)
	req, _ := http.NewRequest("POST", "/ric/v1/xapps", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseData(t, response, http.StatusInternalServerError, false)
}

func TestDeleteXappListReturnsError(t *testing.T) {
	helmError = errors.New("Internal error")

	req, _ := http.NewRequest("DELETE", "/ric/v1/xapps/invalidXappName", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusInternalServerError, response.Code)
}

// Helper functions
type fn func(w http.ResponseWriter, r *http.Request)

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	vars := map[string]string{
		"id": "1",
	}
	req = mux.SetURLVars(req, vars)

	x.router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func checkResponseData(t *testing.T, response *httptest.ResponseRecorder, expectedHttpStatus int, isList bool) {
	expectedData := xapp

	checkResponseCode(t, expectedHttpStatus, response.Code)
	if isList == true {
		jsonResp := []Xapp{}
		json.NewDecoder(response.Body).Decode(&jsonResp)

		if !reflect.DeepEqual(jsonResp[0], expectedData) {
			t.Errorf("handler returned unexpected body: %v", jsonResp)
		}
	} else {
		json.NewDecoder(response.Body).Decode(&xapp)

		if !reflect.DeepEqual(xapp, expectedData) {
			t.Errorf("handler returned unexpected body: got: %v, expected: %v", xapp, expectedData)
		}
	}
}

func generateXapp(name, status, ver, iname, istatus, ip, port string) (x Xapp) {
	x.Name = name
	x.Status = status
	x.Version = ver
	p, _ := strconv.Atoi(port)
	var msgs MessageTypes

	instance := XappInstance{
		Name:       iname,
		Status:     istatus,
		Ip:         ip,
		Port:       p,
		TxMessages: msgs.TxMessages,
		RxMessages: msgs.RxMessages,
	}
	x.Instances = append(x.Instances, instance)

	return
}
