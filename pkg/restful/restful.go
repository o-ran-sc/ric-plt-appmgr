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

package restful

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/restapi"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/restapi/operations"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/restapi/operations/health"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/restapi/operations/xapp"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/valyala/fastjson"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/resthooks"
)

type XappData struct {
	httpendpoint     string
	rmrendpoint      string
	rmrserviceep     string
	status           string
	xappname         string
	xappinstname     string
	xappversion      string
	xappconfigpath   string
	xappdynamiconfig bool
	xappInstance     *models.XappInstance
}

var xappmap = map[string]map[string]*XappData{}

func NewRestful() *Restful {
	r := &Restful{
		rh:    resthooks.NewResthook(true),
		ready: false,
	}
	r.api = r.SetupHandler()
	return r
}

func (r *Restful) Run() {
	server := restapi.NewServer(r.api)
	defer server.Shutdown()
	server.Port = 8080
	server.Host = "0.0.0.0"

	appmgr.Logger.Info("Xapp manager started ... serving on %s:%d\n", server.Host, server.Port)

	go r.symptomdataServer()
	go r.RetrieveApps()
	if err := server.Serve(); err != nil {
		log.Fatal(err.Error())
	}

}

func (r *Restful) RetrieveApps() {
	time.Sleep(5 * time.Second)
	var xlist models.RegisterRequest
	applist := r.rh.GetAppsInSDL()
	if applist != nil {
		appmgr.Logger.Info("List obtained from GetAppsInSDL is %s", *applist)
		newstring := strings.Split(*applist, " ")
		for i, _ := range newstring {
			appmgr.Logger.Debug("Checking for xapp %s", newstring[i])
			if newstring[i] != "" {
				err := json.Unmarshal([]byte(newstring[i]), &xlist)
				if err != nil {
					appmgr.Logger.Error("Error while unmarshalling")
					continue
				}
			} else {
				continue //SDL may have empty item,so need to skip
			}

			xmodel, _ := r.PrepareConfig(xlist, false)
			if xmodel == nil {
				appmgr.Logger.Error("Xapp not found, deleting it from DB")
				r.rh.UpdateAppData(xlist, true)
			}
		}
	}

}

func (r *Restful) SetupHandler() *operations.AppManagerAPI {
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		appmgr.Logger.Error(err.Error())
		os.Exit(1)
	}
	api := operations.NewAppManagerAPI(swaggerSpec)

	// URL: /ric/v1/health
	api.HealthGetHealthAliveHandler = health.GetHealthAliveHandlerFunc(
		func(params health.GetHealthAliveParams) middleware.Responder {
			return health.NewGetHealthAliveOK()
		})

	api.HealthGetHealthReadyHandler = health.GetHealthReadyHandlerFunc(
		func(params health.GetHealthReadyParams) middleware.Responder {
			return health.NewGetHealthReadyOK()
		})

	// URL: /ric/v1/subscriptions
	api.GetSubscriptionsHandler = operations.GetSubscriptionsHandlerFunc(
		func(params operations.GetSubscriptionsParams) middleware.Responder {
			return operations.NewGetSubscriptionsOK().WithPayload(r.rh.GetAllSubscriptions())
		})

	api.GetSubscriptionByIDHandler = operations.GetSubscriptionByIDHandlerFunc(
		func(params operations.GetSubscriptionByIDParams) middleware.Responder {
			if result, found := r.rh.GetSubscriptionById(params.SubscriptionID); found {
				return operations.NewGetSubscriptionByIDOK().WithPayload(&result)
			}
			return operations.NewGetSubscriptionByIDNotFound()
		})

	api.AddSubscriptionHandler = operations.AddSubscriptionHandlerFunc(
		func(params operations.AddSubscriptionParams) middleware.Responder {
			return operations.NewAddSubscriptionCreated().WithPayload(r.rh.AddSubscription(*params.SubscriptionRequest))
		})

	api.ModifySubscriptionHandler = operations.ModifySubscriptionHandlerFunc(
		func(params operations.ModifySubscriptionParams) middleware.Responder {
			if _, ok := r.rh.ModifySubscription(params.SubscriptionID, *params.SubscriptionRequest); ok {
				return operations.NewModifySubscriptionOK()
			}
			return operations.NewModifySubscriptionBadRequest()
		})

	api.DeleteSubscriptionHandler = operations.DeleteSubscriptionHandlerFunc(
		func(params operations.DeleteSubscriptionParams) middleware.Responder {
			if _, ok := r.rh.DeleteSubscription(params.SubscriptionID); ok {
				return operations.NewDeleteSubscriptionNoContent()
			}
			return operations.NewDeleteSubscriptionBadRequest()
		})

	// URL: /ric/v1/xapp
	api.XappGetAllXappsHandler = xapp.GetAllXappsHandlerFunc(
		func(params xapp.GetAllXappsParams) middleware.Responder {
			if result, err := r.GetApps(); err == nil {
				return xapp.NewGetAllXappsOK().WithPayload(result)
			}
			return xapp.NewGetAllXappsInternalServerError()
		})

	// URL: /ric/v1/config
	api.XappGetAllXappConfigHandler = xapp.GetAllXappConfigHandlerFunc(
		func(params xapp.GetAllXappConfigParams) middleware.Responder {
			return xapp.NewGetAllXappConfigOK().WithPayload(r.getAppConfig())
		})

	api.RegisterXappHandler = operations.RegisterXappHandlerFunc(
		func(params operations.RegisterXappParams) middleware.Responder {
			appmgr.Logger.Info("appname is %s", (*params.RegisterRequest.AppName))
			appmgr.Logger.Info("endpoint is %s", (*params.RegisterRequest.HTTPEndpoint))
			appmgr.Logger.Info("rmrendpoint is %s", (*params.RegisterRequest.RmrEndpoint))
			if result, err := r.RegisterXapp(*params.RegisterRequest); err == nil {
				go r.rh.PublishSubscription(*result, models.EventTypeDeployed)
				return operations.NewRegisterXappCreated()
			}
			return operations.NewRegisterXappBadRequest()
		})

	api.DeregisterXappHandler = operations.DeregisterXappHandlerFunc(
		func(params operations.DeregisterXappParams) middleware.Responder {
			appmgr.Logger.Info("appname is %s", (*params.DeregisterRequest.AppName))
			if result, err := r.DeregisterXapp(*params.DeregisterRequest); err == nil {
				go r.rh.PublishSubscription(*result, models.EventTypeUndeployed)
				return operations.NewDeregisterXappNoContent()
			}
			return operations.NewDeregisterXappBadRequest()
		})

	return api
}

func httpGetXAppsconfig(url string) *string {
	appmgr.Logger.Info("Invoked httprestful.httpGetXApps: " + url)
	resp, err := http.Get(url)
	if err != nil {
		appmgr.Logger.Error("Error while querying config to Xapp: ", err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var data XappConfigList
		appmgr.Logger.Info("http client raw response: %v", resp)
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			appmgr.Logger.Error("Json decode failed: " + err.Error())
			return nil
		}
		//data[0] assuming only for one app
		str := fmt.Sprintf("%v", data[0].Config)
		appmgr.Logger.Info("HTTP BODY: %v", str)

		resp.Body.Close()
		return &str
	} else {
		appmgr.Logger.Info("httprestful got an unexpected http status code: %v", resp.StatusCode)
		return nil
	}
}

func parseConfig(config *string) *appmgr.RtmData {
	var p fastjson.Parser
	var msgs appmgr.RtmData

	v, err := p.Parse(*config)
	if err != nil {
		appmgr.Logger.Info("fastjson.Parser for failed: %v", err)
		return nil
	}

	if v.Exists("rmr") {
		for _, m := range v.GetArray("rmr", "txMessages") {
			msgs.TxMessages = append(msgs.TxMessages, strings.Trim(m.String(), `"`))
		}

		for _, m := range v.GetArray("rmr", "rxMessages") {
			msgs.RxMessages = append(msgs.RxMessages, strings.Trim(m.String(), `"`))
		}

		for _, m := range v.GetArray("rmr", "policies") {
			if val, err := strconv.Atoi(strings.Trim(m.String(), `"`)); err == nil {
				msgs.Policies = append(msgs.Policies, int64(val))
			}
		}
	} else {
		for _, p := range v.GetArray("messaging", "ports") {
			appmgr.Logger.Info("txMessages=%v, rxMessages=%v", p.GetArray("txMessages"), p.GetArray("rxMessages"))
			for _, m := range p.GetArray("txMessages") {
				msgs.TxMessages = append(msgs.TxMessages, strings.Trim(m.String(), `"`))
			}

			for _, m := range p.GetArray("rxMessages") {
				msgs.RxMessages = append(msgs.RxMessages, strings.Trim(m.String(), `"`))
			}

			for _, m := range p.GetArray("policies") {
				if val, err := strconv.Atoi(strings.Trim(m.String(), `"`)); err == nil {
					msgs.Policies = append(msgs.Policies, int64(val))
				}
			}
		}
	}
	return &msgs
}

func (r *Restful) RegisterXapp(params models.RegisterRequest) (xapp *models.Xapp, err error) {
	return r.PrepareConfig(params, true)
}

func (r *Restful) DeregisterXapp(params models.DeregisterRequest) (xapp *models.Xapp, err error) {
	var registeredlist models.RegisterRequest
	registeredlist.AppName = params.AppName
	registeredlist.AppInstanceName = params.AppInstanceName
	if _, found := xappmap[*params.AppName]; found {
		var x models.Xapp
		x.Instances = append(x.Instances, xappmap[*params.AppName][*params.AppInstanceName].xappInstance)
		registeredlist.HTTPEndpoint = &xappmap[*params.AppName][*params.AppInstanceName].httpendpoint
		delete(xappmap[*params.AppName], *params.AppInstanceName)
		if len(xappmap[*params.AppName]) == 0 {
			delete(xappmap, *params.AppName)
		}
		r.rh.UpdateAppData(registeredlist, true)
		return &x, nil
	} else {
		appmgr.Logger.Error("XApp Instance %v Not Found", *params.AppName)
		return nil, errors.New("XApp Instance Not Found")
	}
}

func (r *Restful) PrepareConfig(params models.RegisterRequest, updateflag bool) (xapp *models.Xapp, err error) {
	maxRetries := 5
	configPresent := false
	var xappconfig *string
	appmgr.Logger.Info("http endpoint is %s", *params.HTTPEndpoint)
	for i := 1; i <= maxRetries; i++ {
		if params.Config != "" {
			appmgr.Logger.Info("Getting config during xapp register: %v", params.Config)
			xappconfig = &params.Config
			configPresent = true
		} else {
			appmgr.Logger.Info("Getting config from xapp:")
			xappconfig = httpGetXAppsconfig(fmt.Sprintf("http://%s%s", *params.HTTPEndpoint, params.ConfigPath))
		}

		if xappconfig != nil {
			data := parseConfig(xappconfig)
			if data != nil {
				var xapp models.Xapp

				xapp.Name = params.AppName
				xapp.Version = params.AppVersion
				//xapp.Status = params.Status

				r.rh.UpdateAppData(params, updateflag)
				return r.FillInstanceData(params, &xapp, *data, configPresent)
				break
			} else {
				appmgr.Logger.Error("No Data from xapp")
			}
			if configPresent == true {
				break
			}
		}
		appmgr.Logger.Info("Retrying query configuration from xapp, try no : %d", i)
		time.Sleep(4 * time.Second)
	}
	return nil, errors.New("Unable to get configmap after 5 retries")
}

func (r *Restful) FillInstanceData(params models.RegisterRequest, xapp *models.Xapp, rtData appmgr.RtmData, configFlag bool) (xapps *models.Xapp, err error) {

	endPointStr := strings.Split(*params.RmrEndpoint, ":")
	var x models.XappInstance
	x.Name = params.AppInstanceName
	//x.Status = strings.ToLower(params.Status)
	x.Status = "deployed"
	//x.IP = endPointStr[0]
	x.IP = fmt.Sprintf("service-ricxapp-%s-rmr.ricxapp", *params.AppInstanceName)
	x.Port, _ = strconv.ParseInt(endPointStr[1], 10, 64)
	x.TxMessages = rtData.TxMessages
	x.RxMessages = rtData.RxMessages
	x.Policies = rtData.Policies
	xapp.Instances = append(xapp.Instances, &x)
	rmrsrvname := fmt.Sprintf("service-ricxapp-%s-rmr.ricxapp:%s", *params.AppInstanceName, x.Port)

	a := &XappData{httpendpoint: *params.HTTPEndpoint,
		rmrendpoint:      *params.RmrEndpoint,
		rmrserviceep:     rmrsrvname,
		status:           "deployed",
		xappname:         *params.AppName,
		xappversion:      params.AppVersion,
		xappinstname:     *params.AppInstanceName,
		xappconfigpath:   params.ConfigPath,
		xappdynamiconfig: configFlag,
		xappInstance:     &x}

	if _, ok := xappmap[*params.AppName]; ok {
		xappmap[*params.AppName][*params.AppInstanceName] = a
		appmgr.Logger.Info("appname already present, %v", xappmap[*params.AppName])
	} else {
		xappmap[*params.AppName] = make(map[string]*XappData)
		xappmap[*params.AppName][*params.AppInstanceName] = a
		appmgr.Logger.Info("Creating app instance, %v", xappmap[*params.AppName])
	}

	return xapp, nil

}

func (r *Restful) GetApps() (xapps models.AllDeployedXapps, err error) {
	xapps = models.AllDeployedXapps{}
	for _, v := range xappmap {
		var x models.Xapp
		for i, j := range v {
			x.Status = j.status
			x.Name = &j.xappname
			x.Version = j.xappversion
			appmgr.Logger.Info("Xapps details currently in map Appname = %v,rmrendpoint = %v,Status = %v", i, j.rmrendpoint, j.status)
			x.Instances = append(x.Instances, j.xappInstance)
		}
		xapps = append(xapps, &x)
	}

	return xapps, nil

}

func (r *Restful) getAppConfig() (configList models.AllXappConfig) {
	for _, v := range xappmap {
		namespace := "ricxapp" //Namespace hardcoded, to be removed later
		for _, j := range v {
			var activeConfig interface{}
			if j.xappdynamiconfig {
				continue
			}
			xappconfig := httpGetXAppsconfig(fmt.Sprintf("http://%s%s", j.httpendpoint, j.xappconfigpath))

			if xappconfig == nil {
				appmgr.Logger.Info("config not found for %s", &j.xappname)
				continue
			}
			json.Unmarshal([]byte(*xappconfig), &activeConfig)

			c := models.XAppConfig{
				Metadata: &models.ConfigMetadata{XappName: &j.xappname, Namespace: &namespace},
				Config:   activeConfig,
			}
			configList = append(configList, &c)

		}

	}
	return
}

func (r *Restful) symptomdataServer() {
	http.HandleFunc("/ric/v1/symptomdata", func(w http.ResponseWriter, req *http.Request) {
		d, _ := r.GetApps()
		xappData := struct {
			XappList         models.AllDeployedXapps `json:"xappList"`
			ConfigList       models.AllXappConfig    `json:"configList"`
			SubscriptionList models.AllSubscriptions `json:"subscriptionList"`
		}{
			d,
			r.getAppConfig(),
			r.rh.GetAllSubscriptions(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=platform/apps_info.json")
		w.WriteHeader(http.StatusOK)
		resp, _ := json.MarshalIndent(xappData, "", "    ")
		w.Write(resp)
	})

	http.ListenAndServe(":8081", nil)
}
