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
	"errors"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

// API functions

func (m *XappManager) Initialize(h Helmer, cm ConfigMapper) {
	m.cm = cm
	m.helm = h
	m.helm.SetCM(cm)

	m.router = mux.NewRouter().StrictSlash(true)

	resources := []Resource{
		{"GET", "/ric/v1/health/alive", m.getHealthStatus},
		{"GET", "/ric/v1/health/ready", m.getHealthStatus},

		{"GET", "/ric/v1/xapps", m.getAllXapps},
		{"GET", "/ric/v1/xapps/{name}", m.getXappByName},
		{"GET", "/ric/v1/xapps/{name}/instances/{id}", m.getXappInstanceByName},
		{"POST", "/ric/v1/xapps", m.deployXapp},
		{"DELETE", "/ric/v1/xapps/{name}", m.undeployXapp},

		{"GET", "/ric/v1/subscriptions", m.getSubscriptions},
		{"POST", "/ric/v1/subscriptions", m.addSubscription},
		{"GET", "/ric/v1/subscriptions/{id}", m.getSubscription},
		{"DELETE", "/ric/v1/subscriptions/{id}", m.deleteSubscription},
		{"PUT", "/ric/v1/subscriptions/{id}", m.updateSubscription},

		{"GET", "/ric/v1/config", m.getConfig},
		{"POST", "/ric/v1/config", m.createConfig},
		{"PUT", "/ric/v1/config", m.updateConfig},
		{"DELETE", "/ric/v1/config", m.deleteConfig},
		{"DELETE", "/ric/v1/config/{name}", m.deleteSingleConfig},
	}

	for _, resource := range resources {
		handler := Logger(resource.HandlerFunc)
		//handler = m.serviceChecker(handler)
		m.router.Methods(resource.Method).Path(resource.Url).Handler(handler)
	}

	go m.finalize(h)
}

func (m *XappManager) finalize(h Helmer) {
	m.sd = SubscriptionDispatcher{}
	m.sd.Initialize()

	m.helm.Initialize()

	m.notifyClients()
	m.ready = true
}

func (m *XappManager) serviceChecker(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/ric/v1/health/alive" || m.ready == true {
			inner.ServeHTTP(w, r)
		} else {
			respondWithJSON(w, http.StatusServiceUnavailable, nil)
		}
	})
}

// API: XAPP handlers
func (m *XappManager) getHealthStatus(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, nil)
}

func (m *XappManager) Run() {
	host := viper.GetString("local.host")
	if host == "" {
		host = ":8080"
	}
	log.Printf("Xapp manager started ... serving on %s\n", host)

	log.Fatal(http.ListenAndServe(host, m.router))
}

func (m *XappManager) getXappByName(w http.ResponseWriter, r *http.Request) {
	xappName, ok := getResourceId(r, w, "name")
	if ok != true {
		return
	}

	if xapp, err := m.helm.Status(xappName); err == nil {
		respondWithJSON(w, http.StatusOK, xapp)
	} else {
		respondWithError(w, http.StatusNotFound, err.Error())
	}
}

func (m *XappManager) getXappInstanceByName(w http.ResponseWriter, r *http.Request) {
	xappName, ok := getResourceId(r, w, "name")
	if ok != true {
		return
	}

	xapp, err := m.helm.Status(xappName)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	xappInstanceName, ok := getResourceId(r, w, "id")
	if ok != true {
		return
	}

	for _, v := range xapp.Instances {
		if v.Name == xappInstanceName {
			respondWithJSON(w, http.StatusOK, v)
			return
		}
	}
	mdclog(MdclogErr, "Xapp instance not found - url="+r.URL.RequestURI())

	respondWithError(w, http.StatusNotFound, "Xapp instance not found")
}

func (m *XappManager) getAllXapps(w http.ResponseWriter, r *http.Request) {
	xapps, err := m.helm.StatusAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, xapps)
}

func (m *XappManager) deployXapp(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		mdclog(MdclogErr, "No xapp data found in request body - url="+r.URL.RequestURI())
		respondWithError(w, http.StatusMethodNotAllowed, "No xapp data!")
		return
	}

	var cm XappDeploy
	if err := json.NewDecoder(r.Body).Decode(&cm); err != nil {
		mdclog(MdclogErr, "Invalid xapp data in request body - url="+r.URL.RequestURI())
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid xapp data!")
		return
	}
	defer r.Body.Close()

	xapp, err := m.helm.Install(cm)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, xapp)

	m.sd.Publish(xapp, EventType("created"))
}

func (m *XappManager) undeployXapp(w http.ResponseWriter, r *http.Request) {
	xappName, ok := getResourceId(r, w, "name")
	if ok != true {
		return
	}

	xapp, err := m.helm.Delete(xappName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

	m.sd.Publish(xapp, EventType("deleted"))
}

// API: resthook handlers
func (m *XappManager) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, m.sd.GetAll())
}

func (m *XappManager) getSubscription(w http.ResponseWriter, r *http.Request) {
	if id, ok := getResourceId(r, w, "id"); ok == true {
		if s, ok := m.sd.Get(id); ok {
			respondWithJSON(w, http.StatusOK, s)
		} else {
			mdclog(MdclogErr, "Subscription not found - url="+r.URL.RequestURI())
			respondWithError(w, http.StatusNotFound, "Subscription not found")
		}
	}
}

func (m *XappManager) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	if id, ok := getResourceId(r, w, "id"); ok == true {
		if _, ok := m.sd.Delete(id); ok {
			respondWithJSON(w, http.StatusNoContent, nil)
		} else {
			mdclog(MdclogErr, "Subscription not found - url="+r.URL.RequestURI())
			respondWithError(w, http.StatusNotFound, "Subscription not found")
		}
	}
}

func (m *XappManager) addSubscription(w http.ResponseWriter, r *http.Request) {
	var req SubscriptionReq
	if r.Body == nil || json.NewDecoder(r.Body).Decode(&req) != nil {
		mdclog(MdclogErr, "Invalid request payload - url="+r.URL.RequestURI())
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	respondWithJSON(w, http.StatusCreated, m.sd.Add(req))
}

func (m *XappManager) updateSubscription(w http.ResponseWriter, r *http.Request) {
	if id, ok := getResourceId(r, w, "id"); ok == true {
		var req SubscriptionReq
		if r.Body == nil || json.NewDecoder(r.Body).Decode(&req) != nil {
			mdclog(MdclogErr, "Invalid request payload - url="+r.URL.RequestURI())
			respondWithError(w, http.StatusMethodNotAllowed, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		if s, ok := m.sd.Update(id, req); ok {
			respondWithJSON(w, http.StatusOK, s)
		} else {
			mdclog(MdclogErr, "Subscription not found - url="+r.URL.RequestURI())
			respondWithError(w, http.StatusNotFound, "Subscription not found")
		}
	}
}

func (m *XappManager) notifyClients() {
	xapps, err := m.helm.StatusAll()
	if err != nil {
		mdclog(MdclogInfo, "Couldn't fetch xapps status information"+err.Error())
		return
	}

	m.sd.notifyClients(xapps, "updated")
}

func (m *XappManager) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg := m.cm.UploadConfig()
	respondWithJSON(w, http.StatusOK, cfg)
}

func (m *XappManager) createConfig(w http.ResponseWriter, r *http.Request) {
	var c XAppConfig
	if parseConfig(w, r, &c) != nil {
		return
	}

	if errList, err := m.cm.CreateConfigMap(c); err != nil {
		if err.Error() != "Validation failed!" {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		} else {
			respondWithJSON(w, http.StatusUnprocessableEntity, errList)
		}
		return
	}
	respondWithJSON(w, http.StatusCreated, c.Metadata)
}

func (m *XappManager) updateConfig(w http.ResponseWriter, r *http.Request) {
	var c XAppConfig
	if parseConfig(w, r, &c) != nil {
		return
	}

	if errList, err := m.cm.UpdateConfigMap(c); err != nil {
		if err.Error() != "Validation failed!" {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		} else {
			respondWithJSON(w, http.StatusInternalServerError, errList)
		}
		return
	}
	respondWithJSON(w, http.StatusOK, c.Metadata)
}

func (m *XappManager) deleteSingleConfig(w http.ResponseWriter, r *http.Request) {
	xappName, ok := getResourceId(r, w, "name")
	if ok != true {
		return
	}

	md := ConfigMetadata{Name: xappName, Namespace: m.cm.getNamespace(""), ConfigName: xappName + "-appconfig"}
	m.delConfig(w, XAppConfig{Metadata: md})
}

func (m *XappManager) deleteConfig(w http.ResponseWriter, r *http.Request) {
	var c XAppConfig
	if parseConfig(w, r, &c) != nil {
		return
	}

	m.delConfig(w, c)
}

func (m *XappManager) delConfig(w http.ResponseWriter, c XAppConfig) {
	if _, err := m.cm.DeleteConfigMap(c); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		response, _ := json.Marshal(payload)
		w.Write(response)
	}
}

func getResourceId(r *http.Request, w http.ResponseWriter, pattern string) (id string, ok bool) {
	if id, ok = mux.Vars(r)[pattern]; ok != true {
		mdclog(MdclogErr, "Couldn't resolve name/id from the request URL")
		respondWithError(w, http.StatusMethodNotAllowed, "Couldn't resolve name/id from the request URL")
		return
	}
	return
}

func parseConfig(w http.ResponseWriter, r *http.Request, req *XAppConfig) error {
	if r.Body == nil || json.NewDecoder(r.Body).Decode(&req) != nil {
		mdclog(MdclogErr, "Invalid request payload - url="+r.URL.RequestURI())
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request payload")
		return errors.New("Invalid payload")
	}
	defer r.Body.Close()

	return nil
}
