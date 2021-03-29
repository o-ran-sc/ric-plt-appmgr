.. This work is licensed under a Creative Commons Attribution 4.0 International License.
.. http://creativecommons.org/licenses/by/4.0
..
.. Copyright (C) 2019 AT&T Intellectual Property


xApp Descriptor JSON Schema
===========================

Introduction
------------

JSON schema is used to describe the attributes and values in the xApp descriptor config-file.json file. The xApp onboarding process verifies the types and values of the xApp parameters in the descriptor. If mismatches are found, xApp onboarding will return failure. The schema file consists of two parts: sections that are static and cannot be changed for different xApp, and xApp specific controls section. When an operator is onboarding an xApp that defines a control section, he/she will provide the controls section schema with together with the xApp descriptor.

The xapp_onboarder will combine the schema files into one.

Control Section Schema
----------------------
.. note:: No control section schema is needed if your xApp doesn't need a controll section in the config-file.json.  

You can submit arbitrary schema for the controls section. However, if the xApp descriptor contains a controls section, you have to provide the correct schema that describes it. If the xApp does not require a control section, you can ignore the control section schema. It is highly recommended to use draft-07 schema. The following is a skeleton schema that you can use

.. code-block::

  {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "$id": "#/controls",
      "type": "object",
      "title": "Controls Section Schema",
      "required": [
          "REQUIRED_ITEM_1",
          "REQUIRED_ITEM_2"
      ],
      "properties": {
          "REQUIRED_ITEM_1": {REQUIRED_ITEM_1_SUB_SCHEMA},
          "REQUIRED_ITEM_2": {REQUIRED_ITEM_2_SUB_SCHEMA}
      }
  }


Embedded JOSN Schema
--------------------
The following JSON schema is provided by the xApp-onboarder. It defines the JSON file structure of the config-file.json file except the control section.
Expand the following link to read more details.

.. toggle-header::
    :header: **embedded JSON schema**

      .. literalinclude:: embedded-schema.json
        :language: JSON
