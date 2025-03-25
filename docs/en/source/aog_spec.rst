==========================
General Byze Specification
==========================

.. include:: global.rst

Versioning
====================

Current version |byze_spec_version|

The versioning follows semantic versioning, basically:

- MAJOR version when incompatible API changes are made
- MINOR version when functionality in a backward compatible manner is added
- PATCH version when backward compatible bug fixes is made

Therefore, the MAJOR version number is part of the URL to access ``Byze APIs``


Byze Service
====================

``Byze Service`` is a service that provides AI capabilities. It can be a chatbot,
a text-to-speech service, a translation service, etc. Each ``Byze Service`` has
its own ``properties`` and ``API``.

* ``properties`` describes the characteristics of the service, such as the model
  it uses, the hybrid policy it employs, etc.

* ``API`` describes the interface of the service, such as the input/output format,
  the HTTP method it uses, etc.

Byze Service Name and Endpoint
-----------------------------------------

Each ``Byze Service`` has its unique name. The name is used to identify the
service and also to form the URL of endpoint to access the service.

Typical names of the ``Byze Service`` include ``chat``, ``audio/text-to-speech``,
``embed`` etc.

``Byze`` so far supports RESTful like API. The ``Endpoint`` of the ``Byze
Service`` includes both the HTTP method and the URL to access the service. The 
URL is formed by the hostname and port (by default is `localhost` and `16688`) 
``Byze`` prefix, the spec version number, and the name of the
service, i.e. ``http://localhost:{port_number}/byze/v{spec_version_number}/services/{service_name}``.

For example, the ``Endpoint`` of the ``chat`` service is ``POST
http://localhost:16688/byze/v0.2/services/chat``.

.. _`Metadata of Byze Service`:

Metadata of Byze Service
-----------------------------------

Metadata of Byze Service is a JSON object that describes the service. It may vary
from service to service. However, there are some common fields that are usually
included in the metadata of the service.

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - Field
     - Value
     - Description
   * - hybrid_policy
     - ``always_remote``, ``always_local``, ``default``
     - The hybrid policy to use
   * - service_providers
     - JSON object listing the service providers for this service at local and 
       remote. e.g., ``{"local": "provider_a", "remote": "provider_b"}``. Here,
       ``provider_a`` and ``provider_b`` are the ids of the service providers to
       reference the `Metadata of Byze Service Provider`_
     - Describing the service providers for this service


Byze Service Provider
=========================

Byze Service Provider actually implements the corresponding Byze Service.

Typically, for a given Byze Service, it may have multiple service providers. For
instance, a ``/chat`` Byze Service may have a local Byze Service Provider running
on the same machine by ollama, and meanwhile have a remote Byze Service Provider
by OpenAI.

Each Byze Service Provider has a unique id. The id is used to identify the
service provider and used to reference the service provider, e.g., in the
metadata of the Byze Service. See `Metadata of Byze Service`_ .

.. _`Metadata of Byze Service Provider`:

Metadata of Byze Service Provider
------------------------------------

Metadata of Byze Service Provider is a JSON object that describes the service
provider. Here are some common fields that most of the service providers have.

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - Field
     - Value
     - Description
   * - desc
     - string
     - A description of the service provider
   * - method
     - string, e.g., ``POST``
     - The HTTP method to use when invoking the service provider
   * - url
     - string
     - The URL to access the service provider. This together with method forms
       the endpoint of the service provider
   * - api_flavor
     - string, e.g., ``byze``, ``openai``, ``ollama``, etc.
     - The API flavor of the service provider. This is used to convert the API
       if needed. See `flavor_conversion`_
   * - supported_response_mode
     -  ``"sync"``, ``"stream"``, or ``["sync", "stream"]``
     - Whether it supports stream mode or not
   * - allow_to_select_model
     - ``true`` or ``false``. Default is ``true``
     - Whether allow to specify the model when invoking the service
   * - models
     - a list of model names. The first is the default one. e.g., 
       ``["model_a", "model_b"]``For service that needs multiple models, this is
       a JSON object with model lists for each model parameter. e.g.,
       ``{"a": ["model_a1", "model_a2"], "b": ["model_b1", "model_b2"]}``
     - The list of available models to select when ``allow_to_select_model`` is true
   * - extra_headers
     - JSON object
     - Anything else need to add to the header when send the request. For
       example, authorization related headers
   * - extra_json_body
     - JSON object
     - Anything else need to add to the request body for this cloud service
   * - custom_properties
     - JSON object
     - Additional non general properties for each service or service provider.

General Data Types
=========================

In the RESTful API of the ``Byze Service``, there are some commonly used data
types. For example, how to represent the text images, etc. We describe these
general types here.


.. _`content_type`:

Content
-----------

Content can appear in JSON body of a lot of requests or responses. Content can be in various types.

Text Content
_____________

The text content can be either a string

.. code-block:: json

    "content": "a plan string content"


Or an object with type specified 

.. code-block:: json

    {
        "type": "text",
        "text": "the text content here"
    }


It can be even more complex by adding additional information

.. code-block:: json

    {
        "type": "text",
        "text": {
            "value": "the text content",
            "annotations": ["tag1"]
        }
    }

Image Content
_____________

The image content could be either a base64-encoded string, for example, two base64-encoded images in a list below.

.. code-block:: json

    "images": ["iVBORw0KGgoAAAAN...", "YAAADBPx+VAA..."]


Or it could be an object with URL 

.. code-block:: json

    {
        "type": "image_url",
        "image_url": {
            "url": "http://a.com/b.jpg"
        } 
    }


or base64-encoded string

.. code-block:: json

    {
        "type": "image":
        "image": "iVBORw0KGgoAAAAN..."
    }


OpenAI also provides ```image_file```, which will be in Byze's future roadmap.


Array Content
_____________

Array content is simply a list, which contains content defined above.

.. _`message_type`:

Messages
---------------------------

There are several type of messages each has its own fields. So far, we only
support `system` ,`user` , `assistant`, `tool` message type. 

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 15 55

   * - Field
     - Value
     - Required
     - Description
   * - role
     - ``system``, ``user``, ``assistant``, or ``tool``
     - required
     -  
   * - content
     - See `content_type`_
     - required
     - 
   * - tool_call_id
     - a string id
     - optional, required if role is tool
     - Tool call that this message is responding to. This is similar as OpenAI
       function calling capabilities in its `chat/completions`


.. _`Message Example`:

Message Example
________________________________

.. code-block:: json

    "messages": [
        {
            "role": "system",
            "content": "You are a helpful assistant can do function call and your team is Tom"
        },
        {
            "role": "user",
            "content": "Hello. What's your name?"
        },
        {
            "role": "assistant",
            "content": "Hello. My name is Tom."
        },
        {
            "role": "user",
            "content": "I am looking for some shirts"
        },
        {
            "role": "assistant", 
            "tool_calls": [{
                "id": "call_BEGxtsoiM96M78Y97RFxPRYk", 
                "type": "function", 
                "function": {"name": "search", "arguments": "{'query':'shirts'}"}
            }]
        },
        {
            "tool_call_id": "call_BEGxtsoiM96M78Y97RFxPRYk", 
            "role": "tool", 
            "name": "search", 
            "content": "['shirt1', 'shirt2', 'shirt3']"
        },
        {
            "role": "assistant", 
            "content": "I found some options for shirts:\n\n1. Shirt 1\n2. Shirt 2\n3. Shirt 3\n\nWould you like more details on any of these?"
        }
    ]

Tool
-----------------

Tool is a way to describe external functions so that the LLM can understand and
generate invocation messages. There are three tool related data types.

.. _`Tool Message`: 

Tool Message
________________

Tool message is a kind of message which usually contains the execution result of the function call and is sent back to LLM as part of the conversation or history. 

It needs to have ``tool_call_id`` to indicate which call that this message is
responding to. And the ``content`` usually is the execution result of LLM
required tool call. See above `message_type`_ for details and `Message
Example`_ .

.. _`Tool Description`:

Tool Description
___________________________________

This describes an available function and provides it to LLM. With that, LLM can
decide whether to suggest the application to invoke this function, and with
which kind of arguments. So it should not only have the name and descriptions
about the function, but should have the schema about its parameters.

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 15 15 40

   * - Field
     - Next Level Field
     - Value
     - Required
     - Description
   * - type 
     - 
     - "function"
     - required
     - specify that this is a function type
   * - function
     - 
     - 
     - 
     -
   * - 
     - name
     - string 
     - required
     -
   * - 
     - description
     - string
     - optional
     - A description of what the function does, used by the model to choose when
       and how to call the function.
   * -
     - parameters
     - JSON object
     - optional
     - The parameters the functions accepts, described as a `JSON
       Schema <https://json-schema.org/understanding-json-schema>`_ object

Here is an example while you can find another one in `Message Example`_ .

.. code-block:: json

    {
        "type": "function",
        "function": {
            "name": "get_current_weather",
            "description": "Get the current weather in a given location",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {
                        "type": "string",
                        "description": "The city and state, e.g. San Francisco, CA",
                    },
                    "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]},
                },
                "required": ["location"],
            },
        }
    }


.. _`Tool Call Description`:

Tool Call Description
_______________________

This describes the ask of invoking a function. It is usually returned by the LLM
in `tool_calls` as part of the message, given provided available `Tool
Description`_ .

.. list-table:: 
   :header-rows: 1
   :widths: 10 10 10 10 60

   * - Field
     - Next Level Field
     - Value
     - Required
     - Description
   * - id 
     - 
     - string
     - required
     - ID of the tool call
   * - type
     - 
     - function
     - required
     - 
   * - function
     - 
     - 
     - 
     -
   * - 
     - name
     - string 
     - required
     - the name of the function to call
   * - 
     - arguments
     - string
     - optional
     - The arguments to call the function with, as generated by the model in
       JSON format. Note that the model does not always generate valid JSON, and
       may hallucinate parameters not defined by your function schema. Validate
       the arguments in your code before calling your function.

This is an example extracted from aobve `Message Example`_ .

.. code-block:: json

    {
        "id": "call_BEGxtsoiM96M78Y97RFxPRYk", 
        "type": "function", 
        "function": {
            "name": "search", 
            "arguments": "{'query':'shirts'}"
        }
    }

Fields of Header and JSON body for Services
======================================================

The services are invoked by HTTP, and Byze services require ``JSON`` format for
both requests and responses. 

Required and Optional Fields
------------------------------

The fields of the header of requests / responses, and the fields of the JSON
body of requests / responses, are either required or optional.

Required and Optional Fields for Requests
___________________________________________

Required fields should be supplied when make the call, i.e. send the request.
They are usually the minimum set of the information that all potential service
providers will use to actually run the service.

Optional fields can be supplied by the application, but there is no guarantee
that the underlying actual service provider supports them. 

**NOTE** that if an optional field is not supported by underlying actual service
provider, it will be removed from the request during the conversion by Byze. This
might result in certain information loss, but avoids unintentional parameters
are sent to service provider. This behavior is up-to-change in future spec. More
details in `flavor_conversion`_ .


.. list-table:: 
   :header-rows: 1
   :widths: 20 30 50

   * - Type of Fields in Request
     - Should Provide when Invoke
     - Will Be Accepted by Underlying Service Provider
   * - required
     - | Yes. 
       | Must be provided unless that field has default value
     - | Yes. 
       | The actual service provider will receive these fields and use them
   * - optional
     - | No. 
       | The caller are free to supply these fields or not
     - | Depends.
       | Only the fields that are supported by the actual Service Provider will 
         be passed to it. Other fields are filtered and ignored.

Required and Optional Fields for Responses
____________________________________________

Required fields defined for response will always be provided by Byze.

However, whether an optional field will appear in the returned response, depends
on the actual Service Provider which processes the corresponding request. 

Unlike the request, the additional fields not defined in Byze but returned by
Service Provider, will still be remained there and sent to the application
without being filtered out. 

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - Type of Fields in Response
     - Will appear in the Response returned  by Underlying Service Provider
     - Will appear in the final Response returned by Byze
   * - required
     - | Unsure. 
       | The response from underlying Service Provider may have the related 
         information but not put in this field.
     - | Yes. 
       | Byze will convert and ensure the field is presented and the
         application is safe to access this field
   * - optional
     - Unsure.
     - | Unsure.
       | The final response returned by Byze only contain it if it is provided
         by the actual Service Provider

.. _`Common Fields of Byze Services`:

Common Fields of Byze Services
==================================

Although each service may have its own specific fields in header and JSON body
of its requests / responses. Some fields are quite common and appear in most the
service requests / responses.

.. _`Common Fields in Header of Request`:

Common Fields in Header of Request
--------------------------------------

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 10 65

   * - Field
     - Value
     - Required
     - Description
   * - Content-Type
     - application/json
     - required
     - Unless specially specified, most of the API endpoints only accept JSON format

Most of the APIs take JSON as the input data so has this ``Content-Type`` in
header. However, for binary inputs, such as images or audio files etc., APIs may
use form data (e.g. ``multipart/form-data`` in ``Content-Type``)

**NOTE** that unlike invoking cloud service, the ``Authorization`` header is not
always required since it may invoke the Byze Service locally. However, such
header can be added by Byze later based on the information provided in
configuration or application invocation. 

.. _`Common Fields in Request Body`:

Common Fields in Request Body
---------------------------------

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 10 60

   * - JSON Field
     - Value
     - Required
     - Description
   * - model
     - string or JSON object
     - optional
     - The model(s) to be used for this service. If not provided, the default
       model will be used. If the service does not support selecting model, this
       field will be ignored. See ``allow_to_select_model`` in 
       `Metadata of Byze Service Provider`_. Some services may have multiple
       models, so the application may pass in a JSON object to specify the
       model name for each model to be used inside the service.
   * - stream
     - ``true`` or ``false``. Default is ``false``
     - required
     - Whether to use stream mode or not. If not provided, the default mode will
       be used. See ``supported_response_mode`` in `Metadata of Byze Service Provider`_
   * - hybrid_policy
     - ``always_remote``, ``always_local``, ``default``
     - optional
     - The hybrid policy to use. If not provided, the ``default`` policy will be
       used. See ``hybrid_policy`` in `Metadata of Byze Service`_
   * - remote_service_provider
     - string, or JSON object as of `Metadata of Byze Service Provider`_
     - optional
     - By default, if Byze decides to use the remote Service Provider to serve
       this invocation (based on hybrid_policy etc.), it will use platform
       defined default remote service provider for this service. However,
       application may overwrite this particular for current invocation, by
       using this field in the request to ensure this remote service provider is
       used when Byze decides to use the remote. The field needs to be string
       which is the ID of a predefined Service Provider. Or it can be a JSON
       object which is the `Metadata of Byze Service Provider`_
   * - keep_alive
     - timespan, e.g. ``5m`` as the default
     - optional
     - hints for underlying services (e.g. ``ollama``) to control how long the
       local model stay in memory



.. _`Common Fields in Response Body`:

Common Fields in Response Body
---------------------------------

In a lot of occasions, the returned response is a JSON object. In addition to
make conversions if needed, (See `flavor_conversion`_), Byze will also add a
field ``byze`` in the returned JSON object. The value of this ``byze`` field is a
JSON object with the following fields.

.. list-table:: 
   :header-rows: 1
   :widths: 20 80

   * - Field under ``byze`` in Response
     - Description
   * - received_request_at
     - The timestamp when the invocation request is received by Byze
   * - received_response_at
     - The timestamp when Byze receives the response from the actual Service
       Provider
   * - served_by
     - The URL of the Service Provider that actually serves this invocation
   * - served_by_api_flavor
     - The API flavor of the Service Provider that actually serves this invocation
   * - model
     - The model(s) actually invoked. This is optional and not always provided
