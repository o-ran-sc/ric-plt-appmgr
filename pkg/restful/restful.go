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
	"log"
	"os"
	"time"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/models"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/restapi"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/restapi/operations"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/restapi/operations/health"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/restapi/operations/xapp"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"

	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/appmgr"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/cm"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/helm"
	"gerrit.oran-osc.org/r/ric-plt/appmgr/pkg/resthooks"
)

func NewRestful() *Restful {
	r := &Restful{
		helm:  helm.NewHelm(),
		cm:    cm.NewCM(),
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

	go r.NotifyClients()
	if err := server.Serve(); err != nil {
		log.Fatal(err.Error())
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
			if result, err := r.helm.StatusAll(); err == nil {
				return xapp.NewGetAllXappsOK().WithPayload(result)
			}
			return xapp.NewGetAllXappsInternalServerError()
		})

	api.XappListAllXappsHandler = xapp.ListAllXappsHandlerFunc(
		func(params xapp.ListAllXappsParams) middleware.Responder {
			if result := r.helm.SearchAll(); err == nil {
				return xapp.NewListAllXappsOK().WithPayload(result)
			}
			return xapp.NewListAllXappsInternalServerError()
		})

	api.XappGetXappByNameHandler = xapp.GetXappByNameHandlerFunc(
		func(params xapp.GetXappByNameParams) middleware.Responder {
			if result, err := r.helm.Status(params.XAppName); err == nil {
				return xapp.NewGetXappByNameOK().WithPayload(&result)
			}
			return xapp.NewGetXappByNameNotFound()
		})

	api.XappGetXappInstanceByNameHandler = xapp.GetXappInstanceByNameHandlerFunc(
		func(params xapp.GetXappInstanceByNameParams) middleware.Responder {
			if result, err := r.helm.Status(params.XAppName); err == nil {
				for _, v := range result.Instances {
					if *v.Name == params.XAppInstanceName {
						return xapp.NewGetXappInstanceByNameOK().WithPayload(v)
					}
				}
			}
			return xapp.NewGetXappInstanceByNameNotFound()
		})

	api.XappDeployXappHandler = xapp.DeployXappHandlerFunc(
		func(params xapp.DeployXappParams) middleware.Responder {
			if result, err := r.helm.Install(*params.XappDescriptor); err == nil {
				go r.PublishXappCreateEvent(params)
				return xapp.NewDeployXappCreated().WithPayload(&result)
			}
			return xapp.NewUndeployXappInternalServerError()
		})

	api.XappUndeployXappHandler = xapp.UndeployXappHandlerFunc(
		func(params xapp.UndeployXappParams) middleware.Responder {
			if result, err := r.helm.Delete(params.XAppName); err == nil {
				go r.PublishXappDeleteEvent(result)
				return xapp.NewUndeployXappNoContent()
			}
			return xapp.NewUndeployXappInternalServerError()
		})

	// URL: /ric/v1/config
	api.XappGetAllXappConfigHandler = xapp.GetAllXappConfigHandlerFunc(
		func(params xapp.GetAllXappConfigParams) middleware.Responder {
			return xapp.NewGetAllXappConfigOK().WithPayload(r.cm.UploadConfigAll())
		})

	api.XappGetConfigElementHandler = xapp.GetConfigElementHandlerFunc(
		func(params xapp.GetConfigElementParams) middleware.Responder {
			return xapp.NewGetConfigElementOK().WithPayload(r.cm.UploadConfigElement(params.Element))
		})

	api.XappModifyXappConfigHandler = xapp.ModifyXappConfigHandlerFunc(
		func(params xapp.ModifyXappConfigParams) middleware.Responder {
			result, err := r.cm.UpdateConfigMap(*params.XAppConfig)
			if err != nil {
				if err.Error() != "Validation failed!" {
					return xapp.NewModifyXappConfigInternalServerError()
				} else {
					return xapp.NewModifyXappConfigUnprocessableEntity()
				}
			}
			r.rh.PublishSubscription(models.Xapp{}, models.EventTypeModified)
			return xapp.NewModifyXappConfigOK().WithPayload(result)
		})

	return api
}

func (r *Restful) NotifyClients() {
	r.helm.Initialize()
	if xapps, err := r.helm.StatusAll(); err == nil {
		r.rh.NotifyClients(xapps, models.EventTypeRestarted)
		r.ready = true
	}
}

func (r *Restful) PublishXappCreateEvent(params xapp.DeployXappParams) {
	name := *params.XappDescriptor.XappName
	if params.XappDescriptor.ReleaseName != "" {
		name = params.XappDescriptor.ReleaseName
	}

	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(5) * time.Second)
		if result, _ := r.helm.Status(name); result.Instances != nil {
			r.rh.PublishSubscription(result, models.EventTypeDeployed)
			break
		}
	}
}

func (r *Restful) PublishXappDeleteEvent(xapp models.Xapp) {
	r.rh.PublishSubscription(xapp, models.EventTypeUndeployed)
}
