.. This work is licensed under a Creative Commons Attribution 4.0 International License.
.. http://creativecommons.org/licenses/by/4.0
..
.. Copyright (C) 2019 AT&T Intellectual Property


Step-by-step xApp Descriptor Design Example
===========================================

Step 1: Gather Information
--------------------------
Collect the information about the following items

* xApp name
* Version of the xApp descriptor
* Details of the xApp containers
* (Optional) xApp-specific configuration parameters
* (Optional) Metrics produced
* (Optional) Ports for messaging
* (Optional) Liveness and readiness probes methods

Step 2: Download config-file.json skeleton
------------------------------------------
Download the config-file.json file :download:`here <config-file.json>`

Step 3: Change xapp_name and version
------------------------------------
In the config-file.json file, change the values for "xapp_name" and "version"

.. note:: If xapp-onboarder is configured with ALLOW_REDEPLOY=False, you cannot reuse the same version number between onboarding.

Step 4: Fill in the container information
-----------------------------------------
The container section in the config-file.json is a list of container properties structure. For each container, give it a unique name. Specify the docker image registry, docker image name, and docker iamge tag. Optionally. Please make sure that the docker registry is accessible from the RIC platform instance to which the xApp will be deployed. If you want to specify the contianer entry point, you can specify the command used to start the container.


 
