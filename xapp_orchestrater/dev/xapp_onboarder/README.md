xApp Onboarder
==============

xApp onboarder onboards xApp to the near-rt RIC platform. The operators provides the xApp descriptors and their schemas, the xApp onboarder generates the xApp helm charts dynamically.

## Install xapp_onboarder

Run  [pip](https://pip.pypa.io/en/stable/) to install xapp_onboarder.

```bash
pip install xapp_onboarder
```

## Prerequisite Requirements
A helm chart repo is needed to store the xApp helm charts. You can use [chartmuseum](https://github.com/helm/chartmuseum) for this purpose.
 
Environment variables:
* **FLASK_SERVER_NAME**: Address that the xapp_onboarder is listening on. Default http://0.0.0.0:8888
* **CHART_REPO_URL**: helm chart repo URL. Default http://0.0.0.0:8080

## Configurations
Environment variables:
* **CHART_WORKSPACE_PATH**: Temporary directory that will store the xApp helm chart artifacts. Default /tmp/xapp_onboarder
* **CHART_WORKSPACE_SIZE**: Size limit of the temporary directory. Default 500MB
* **ALLOW_REDEPLOY**: Enable or disable redeploying of xApp helm charts. Default True
* **HTTP_TIME_OUT**: Timeout of all http requests. Default 10 
* **HTTP_RETRY**: Number of retry xapp_onboarder will use for the http requests. Default 3

## Run the API server
```bash
xapp_onboarder
```
## Run the CLI tool
```bash
cli
```
