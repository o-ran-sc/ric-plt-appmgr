# RIC xApp Manager

Provides a flexible and secure way for deploying and managing various RIC xApp applications.

## Communication Interfaces (draft for R0)
* Northbound (External)
  * RESTful API
* Southbound (internal)
  * Helm  (Package manager for Kubernetes)

## REST services for XApp managements
```sh
Action                      URL                                 Method

Deploy                      /ric/v1/xapps                       POST
Undeploy                    /ric/v1/xapps/{xappName}            DELETE
Query Xapp Status           /ric/v1/xapps/{xappName}            GET
Query Xapp Instance Status  /ric/v1/xapps/instances/{xappName}  GET
Query All Xapp Status       /ric/v1/xapps                       GET
Health Check                /ric/v1/health                      GET
```

## REST services for subscriptions (resthooks)
```sh
Action                      URL                                 Method

Add A Subscription          /ric/v1/subscriptions               POST
Update A Subscription       /ric/v1/subscriptions/{id}          PUT
Delete A Subscription       /ric/v1/subscriptions/{id}          DELETE
Get A Subscription          /ric/v1/subscriptions               GET
Get All Subscriptions       /ric/v1/subscriptions/{id}          GET
```

## Used RIC platform services 
TBD later

## Prerequisites
Make sure that following tools are properly installed and configured
* GO (golang) development and runtime tools
* mdclog (com/log)
* Docker
* Kubernates and related tools (kubectl and helm)
* Xapp Docker repo (either local or remote)
* Xapp Helm charts
* ...

## Building go binary and docker container for xApp Manager
 ```sh
# Run following command. Make sure that mdclog is installed and found in the standard library path
make docker-build
```

## Running xApp Manager unit tests
 ```sh
# Run following command
make test
```

## Running xApp Manager locally
```sh
# Now run the xApp manager
build/appmgr -f config/appmgr.yaml
```

# Running Docker container of xApp manager
```sh
make docker-run
```

# Deploy, undeploying xApps and querying status (using CURL command)
```sh
# Deploy a new xApp instance with the name 'dummy-xapp'
curl -H "Content-Type: application/json" -X POST http://172.17.0.3:8080/ric/v1/xapps -d '{"name": "dummy-xapp"}'
```

```sh
# Query the status of all xApp applications
curl -H "Content-Type: application/json" http://localhost:8080/ric/v1/xapps
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    95  100    95    0     0  95000      0 --:--:-- --:--:-- --:--:-- 95000
[
    {
        "name": "dummy-xapp",
        "status": "DEPLOYED",
        "version": "1.0",
        "instances": [
            {
                "name": "dummy-xapp-8984fc9fd-8jq9q",
                "status": "Running",
                "ip": "10.99.213.161",
                "port": 80,
                "txMessages": "[]",
                "rxMessages": "[]"
            },
            {
                "name": "dummy-xapp-8984fc9fd-zq47z",
                "status": "Running",
                "ip": "10.99.213.161",
                "port": 80,
                "txMessages": "[]",
                "rxMessages": "[]"
            },
            {
                "name": "dummy-xapp-8984fc9fd-zzxjj",
                "status": "Running",
                "ip": "10.99.213.161",
                "port": 80,
                "txMessages": "[]",
                "rxMessages": "[]"
            }
        ]
    }
]
```
```sh
# Query the status of a sigle xApp (using the xApp name)
curl -H "Content-Type: application/json"  http://localhost:8080/ric/v1/xapps/dummy-xapp
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    95  100    95    0     0  95000      0 --:--:-- --:--:-- --:--:-- 95000
{
    "name": "dummy-xapp",
    "status": "DEPLOYED",
    "version": "1.0",
    "instances": [
        {
            "name": "dummy-xapp-8984fc9fd-8jq9q",
            "status": "Running",
            "ip": "10.99.213.161",
            "port": 80,
            "txMessages": "[]",
            "rxMessages": "[]"
        },
        {
            "name": "dummy-xapp-8984fc9fd-zq47z",
            "status": "Running",
            "ip": "10.99.213.161",
            "port": 80,
            "txMessages": "[]",
            "rxMessages": "[]"
        },
        {
            "name": "dummy-xapp-8984fc9fd-zzxjj",
            "status": "Running",
            "ip": "10.99.213.161",
            "port": 80,
            "txMessages": "[]",
            "rxMessages": "[]"
        }
    ]
}
```
```sh
# Query the status of a sigle xApp instance (using the xApp instance name)
curl -H "Content-Type: application/json"  http://localhost:8080/ric/v1/xapps/dummy-xapp
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    95  100    95    0     0  95000      0 --:--:-- --:--:-- --:--:-- 95000
{
    "name": "dummy-xapp-8984fc9fd-8jq9q",
    "status": "Running",
    "ip": "10.99.213.161",
    "port": 80,
    "txMessages": "[]",
    "rxMessages": "[]"
}
```
```sh
# Undeploy xApp by name
curl -H "Content-Type: application/json"  -X DELETE http://localhost:8080/ric/v1/xapps/dummy-xapp
```

# Health Check Probes (using CURL command)
```sh
# Health Check using CURL
curl -H "Content-Type: application/json" http://10.244.1.47:8080/ric/v1/health --verbose
*   Trying 10.244.1.47...
* TCP_NODELAY set
* Connected to 10.244.1.47 (10.244.1.47) port 8080 (#0)
> GET /ric/v1/health HTTP/1.1
> Host: 10.244.1.47:8080
> User-Agent: curl/7.58.0
> Accept: */*
> Content-Type: application/json
> 
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Sun, 24 Mar 2019 11:13:59 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.244.1.47 left intact
```

# Subsciptions: List, create, update and delete (using CURL command)
```sh
# Add a new subscription
curl -H "Content-Type: application/json" http://172.17.0.3:8080/ric/v1/subscriptions -X POST -d '{"maxRetries": 3, "retryTimer": 5, "eventType":"Created", "targetUrl": "http://192.168.0.12:8088/"}'

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   169  100    70  100    99  17500  24750 --:--:-- --:--:-- --:--:-- 56333
{
    "id": "1ILBltYYzEGzWRrVPZKmuUmhwcc",
    "version": 0,
    "eventType": "Created"
}
```
```sh
# List all subscriptions
curl -H "Content-Type: application/json" http://172.17.0.3:8080/ric/v1/subscriptions
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   259  100   259    0     0   252k      0 --:--:-- --:--:-- --:--:--  252k
[
    {
        "id": "1ILBZTtEVVtQmIZnh1OJdBP7bcR",
        "targetUrl": "http://192.168.0.12:8088/",
        "eventType": "Created",
        "maxRetries": 3,
        "retryTimer": 5
    },
    {
        "id": "1ILBltYYzEGzWRrVPZKmuUmhwcc",
        "targetUrl": "http://192.168.0.12:8088/",
        "eventType": "Created",
        "maxRetries": 3,
        "retryTimer": 5
    }
]
```

```sh
# Get a specific subscription by Id
curl -H "Content-Type: application/json" http://172.17.0.3:8080/ric/v1/subscriptions/1ILBZTtEVVtQmIZnh1OJdBP7bcR
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   128  100   128    0     0   125k      0 --:--:-- --:--:-- --:--:--  125k
{
    "id": "1ILBZTtEVVtQmIZnh1OJdBP7bcR",
    "targetUrl": "http://192.168.0.12:8088/",
    "eventType": "Created",
    "maxRetries": 3,
    "retryTimer": 5
}
```

```sh
# Delete a specific subscription by Id
curl -H "Content-Type: application/json" http://172.17.0.3:8080/ric/v1/subscriptions/1ILBZTtEVVtQmIZnh1OJdBP7bcR -X DELETE
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
```

```sh
# Example of subscription notification POSTed to targetUrl provided by the client

{
	"id": "1ILBltYYzEGzWRrVPZKmuUmhwcc",
	"version": 0,
	"eventType": "Created",
	"xapp": {
		"name": "dummy-xapp",
		"status": "DEPLOYED",
		"version": "1.0",
		"instances": [
			{
				"name": "dummy-xapp-8984fc9fd-lh7r2",
				"status": "ContainerCreating",
				"ip": "10.104.73.185",
				"port": 80,
				"txMessages": "[]",
				"rxMessages": "[]"
			},
			{
				"name": "dummy-xapp-8984fc9fd-lzrdk",
				"status": "Pending",
				"ip": "10.104.73.185",
				"port": 80,
				"txMessages": "[]",
				"rxMessages": "[]"
			},
			{
				"name": "dummy-xapp-8984fc9fd-xfjcn",
				"status": "Pending",
				"ip": "10.104.73.185",
				"port": 80,
				"txMessages": "[]",
				"rxMessages": "[]"
			}
		]
	}
}
```

# Using xapp manager CLI (appmgrcli) to manage xapps (deploy, get, undeploy, etc)

Run command *appmgrcli help* for short usage instructions, or read the
script source; the instructions can be found as plain text near the
beginning.

Unlike direct curl commands, using the *appmgrcli* validates some of
the parameters, and there is usually less to type...

The host and port where the xapp manager is running are given by
options *-h* and *-p*, or you can define environment variables
APPMGR_HOST and APPMGR_PORT to specify them (recommended). The
following examples assume they have been specified.

```sh
# Deploy a xapp

$ appmgrcli deploy dummy-xapp
{
    "name": "dummy-xapp",
    "status": "DEPLOYED",
    "version": "1.0",
    "instances": [
        {
            "name": "dummy-xapp-667dfc9bfb-wd5m9",
            "status": "Pending",
            "ip": "",
            "port": 0,
            "txMessages": "",
            "rxMessages": ""
        }
    ]
}

# Undeploy

$ appmgrcli undeploy dummy-xapp
dummy-xapp undeployed

# Add some subscriptions

$ appmgrcli subscriptions add https://kukkuu.reset created 500 600
{
    "id": "1IoQqEI24sPfLkq8prmMqk6Oz1I",
    "version": 0,
    "eventType": "created"
}
$ appmgrcli subscriptions add https://facebook.com all 10 4
{
    "id": "1IoR85ZwgiNiIn82phUR6qJmBvq",
    "version": 0,
    "eventType": "all"
}

# list and delete (also shows using abbreviations):


$ appmgrcli subs list
[
    {
        "id": "1IoQqEI24sPfLkq8prmMqk6Oz1I",
        "targetUrl": "https://kukkuu.reset",
        "eventType": "created",
        "maxRetries": 500,
        "retryTimer": 600
    },
    {
        "id": "1IoR85ZwgiNiIn82phUR6qJmBvq",
        "targetUrl": "https://facebook.com",
        "eventType": "all",
        "maxRetries": 10,
        "retryTimer": 4
    }
]

$ appmgrcli subs del 1IoR85ZwgiNiIn82phUR6qJmBvq
Subscription 1IoR85ZwgiNiIn82phUR6qJmBvq deleted

$ appmgrcli subs list
[
    {
        "id": "1IoQqEI24sPfLkq8prmMqk6Oz1I",
        "targetUrl": "https://kukkuu.reset",
        "eventType": "created",
        "maxRetries": 500,
        "retryTimer": 600
    }
]

```

# Additional info
```sh
Todo
```