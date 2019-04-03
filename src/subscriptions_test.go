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
    "net/http"
    "testing"
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "net"
    "log"
    "fmt"
)

var resp SubscriptionResp

// Test cases
func TestNoSubscriptionsFound(t *testing.T) {
    req, _ := http.NewRequest("GET", "/ric/v1/subscriptions", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
    if body := response.Body.String(); body != "[]" {
        t.Errorf("handler returned unexpected body: got %v want []", body)
    }
}

func TestAddNewSubscription(t *testing.T) {
    payload := []byte(`{"maxRetries": 3, "retryTimer": 5, "eventType":"Created", "targetUrl": "http://localhost:8087/xapps_handler"}`)
    req, _ := http.NewRequest("POST", "/ric/v1/subscriptions", bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusCreated, response.Code)

    json.NewDecoder(response.Body).Decode(&resp)
    if resp.Version != 0 {
        t.Errorf("Creating new subscription failed: %v", resp)
    }
}

func TestGettAllSubscriptions(t *testing.T) {
    req, _ := http.NewRequest("GET", "/ric/v1/subscriptions", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var subscriptions []SubscriptionReq
    json.NewDecoder(response.Body).Decode(&subscriptions)

    verifySubscription(t, subscriptions[0], "http://localhost:8087/xapps_handler", 3, 5, "Created")
}

func TestGetSingleSubscription(t *testing.T) {
    req, _ := http.NewRequest("GET", "/ric/v1/subscriptions/" + resp.Id, nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var subscription SubscriptionReq
    json.NewDecoder(response.Body).Decode(&subscription)

    verifySubscription(t, subscription, "http://localhost:8087/xapps_handler", 3, 5, "Created")
}

func TestUpdateSingleSubscription(t *testing.T) {
    payload := []byte(`{"maxRetries": 11, "retryTimer": 22, "eventType":"Deleted", "targetUrl": "http://localhost:8088/xapps_handler"}`)

    req, _ := http.NewRequest("PUT", "/ric/v1/subscriptions/" + resp.Id, bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var res SubscriptionResp
    json.NewDecoder(response.Body).Decode(&res)
    if res.Version != 0 {
        t.Errorf("handler returned unexpected data: %v", resp)
    }

    // Check that the subscription is updated properly
    req, _ = http.NewRequest("GET", "/ric/v1/subscriptions/" + resp.Id, nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

    var subscription SubscriptionReq
    json.NewDecoder(response.Body).Decode(&subscription)

    verifySubscription(t, subscription, "http://localhost:8088/xapps_handler", 11, 22, "Deleted")
}

func TestDeleteSingleSubscription(t *testing.T) {
    req, _ := http.NewRequest("DELETE", "/ric/v1/subscriptions/" + resp.Id, nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNoContent, response.Code)

    // Check that the subscription is removed properly
    req, _ = http.NewRequest("GET", "/ric/v1/subscriptions/" + resp.Id, nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteSingleSubscriptionFails(t *testing.T) {
    req, _ := http.NewRequest("DELETE", "/ric/v1/subscriptions/invalidSubscriptionId" , nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestAddSingleSubscriptionFailsBodyEmpty(t *testing.T) {
    req, _ := http.NewRequest("POST", "/ric/v1/subscriptions/" + resp.Id , nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

func TestUpdateeSingleSubscriptionFailsBodyEmpty(t *testing.T) {
    req, _ := http.NewRequest("PUT", "/ric/v1/subscriptions/" + resp.Id , nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

func TestUpdateeSingleSubscriptionFailsInvalidId(t *testing.T) {
    payload := []byte(`{"maxRetries": 11, "retryTimer": 22, "eventType":"Deleted", "targetUrl": "http://localhost:8088/xapps_handler"}`)

    req, _ := http.NewRequest("PUT", "/ric/v1/subscriptions/invalidSubscriptionId" + resp.Id, bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestPublishXappAction(t *testing.T) {
    payload := []byte(`{"maxRetries": 3, "retryTimer": 5, "eventType":"Created", "targetUrl": "http://127.0.0.1:8888"}`)
    req, _ := http.NewRequest("POST", "/ric/v1/subscriptions", bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusCreated, response.Code)

    // Create a RestApi server (simulating RM)
    ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, XM!")
    }))

    l, err := net.Listen("tcp", "127.0.0.1:8888")
    if err != nil {
        log.Fatal(err)
    }
    ts.Listener.Close()
    ts.Listener = l
    ts.Start()

    defer ts.Close()

    x.sd.Publish(xapp, EventType("created"))
}

func verifySubscription(t *testing.T, subscription SubscriptionReq, url string, retries int, timer int, event string) {
    if subscription.TargetUrl != url {
        t.Errorf("Unexpected url: got=%s expected=%s", subscription.TargetUrl, url)
    }

    if subscription.MaxRetries != retries {
        t.Errorf("Unexpected retries: got=%d expected=%d", subscription.MaxRetries, retries)
    }

    if subscription.RetryTimer != timer {
        t.Errorf("Unexpected timer: got=%d expected=%d", subscription.RetryTimer, timer)
    }

    if subscription.EventType != event {
        t.Errorf("Unexpected event type: got=%s expected=%s", subscription.EventType, event)
    }
}