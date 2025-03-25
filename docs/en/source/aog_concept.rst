====================================
Byze Key Concepts
====================================

.. include:: global.rst

In this section, we will introduce the key concepts / terminologies of ``Byze``.

Byze Service and Byze Service Providers
=========================================

In ``Byze``, typical general AI services such as ``chat`` and ``text-to-image``
etc. are defined as ``Byze Service``, each with standardized API.

Currently, ``Byze`` supports RESTful API based on HTTP protocol. So the API here
means the endpoint (HTTP verb and URL), and the format of the request and
response through this endpoint for the corresponding service.

``Byze Service Provider``, on the other hand, is the one who actually provides
``Byze Service`` by implementing the API of the service.

This is similar as the concepts of abstract interface and concrete class in OOP
(Object Oriented Programming), i.e. ``Byze Service`` is a kind of abstract
interface, while ``Byze Service Provider`` is the concrete class implementing
that interface.


In practice, one ``Byze Service`` can have multiple ``Byze Service Providers``.
And one ``Byze Service Provider`` can also provide multiple ``Byze Services``.


.. graphviz:: 
   :align: center

   digraph G {
    rankdir=BT
    compound=true
    label = "Byze Services and Service Providers"
    graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
    node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 


    subgraph cluster_byze_service {
        label = "Byze Services"
        color="#dddddd"
        fillcolor="#eeeeee"

        node[fillcolor="#ffffcc"]

        chat[label="chat"]
        rag_summarize[label="rag/summarize"]
        text_to_image[label="text-to-image"]
    }

    subgraph cluster_byze_service_provider {
        label = "Byze Service Providers"
        color="#dddddd"
        fillcolor="#eeeeee"

        node[fillcolor="#eeffcc"]

        provider_a[label="Remote OpenAI Chat"]
        provider_b[label="Local Ollama Chat"]
        provider_c[label="Local Stable Diffusion"]
    }


    edge[arrowhead=onormal]
    {provider_a, provider_b} -> chat
    {provider_a, provider_b} -> rag_summarize
    provider_c -> text_to_image

   }

As illustrated above, both *Remote OpenAI Chat* and *Local Ollama Chat* are
``Byze Service Providers`` for the ``chat`` service. Meanwhile, the
*rag/summarize* service actually relies on the same providers as the ``chat``.



Details of Byze Service
================================

An ``Byze Service`` is defined by the following attributes:

Name
    The name of the service. This serves as the string ID of the ``Byze
    Service``. Thus it should be unique among all services. Examples are
    ``chat``, ``embed``, ``rag/query``, ``text-to-image``,
    ``audio/text-to-speech`` etc.

Endpoints
    This tells the upper level application about how to invoke the service, i.e
    the HTTP verb and the URL. It only needs to specify the partial URL, and a
    prefix will be added according according to real deployment. For example,
    the endpoint for ``chat`` service is ``POST /chat``. In real deployment, the
    full URL path will have the site, port, and byze prefix etc. added, resulting
    in something like ``http://localhost:16688/byze/v0.2/service/chat``.
    
    It is possible to have multiple endpoints for one service while most
    services only have one. An example of multiple endpoints is the ``embed``
    service. It accepts both ``POST`` and ``GET`` request. So have two
    endpoints, i.e. ``POST /embed`` and ``GET /embed``

Request Schema
    The format and information of the request. This tells the upper level
    application about how to prepare the request when invoke the service. As an
    analogy, think of Endpoint as the function name and Request Schema as the
    function signature, defining the number and types of parameters.

    ``Byze`` treats the information of the request as key-value pairs. That
    means, each request consists of a set of ``fields`` and each field has
    certain ``value`` associated with it. How these fields and values are
    encoded in the HTTP Request, depends on the type of request and its content
    type.

    Specifically, for ``GET`` request, the fields and values are encoded as
    query parameters in the URL. For ``POST`` request, the fields and values can
    be encoded in the body of the request. The body can be in different formats,
    such as ``JSON``, ``multipart/form-data``.
    
    ``Byze`` defines the ``required`` and ``optional`` fields for each ``Byze
    Service`` it covers.

Response Schema
    Similar as Request, ``Byze`` defines the ``required`` and ``optional``
    content to be returned by each ``Byze Service``. Again considering above
    analog, think of ``Response Schema`` as the return type of the function.

.. _byze_service_provider_properties:

Properties
    Each ``Byze Service`` also has a set of properties. These properties are used
    to describe the actual service provider that implements the service so the
    value varies from one service provider to another.

    For example, ``chat`` service has the property so called
    ``max_input_tokens``. This property tells the upper level application that
    the maximum number of tokens that can be sent to the underlying service
    provider. Different service providers may have different values for this
    property. 

    One important property of many service providers is ``models``. It tells
    ``Byze`` about its available models. For example, a ``chat`` service provider
    may allow to switch between different models.


API Flavor of Application and Byze Service Provider
========================================================

In :doc:`byze_spec`, ``Byze`` defines its specification for ``Byze Services``
including its Name, Endpoints, Request Schema, Response Schema etc.

However, in reality, as mentioned in :ref:`compatibility_issue`, it is highly
possible that application developers and service providers may not fully adhere
to the ``Byze`` specification.

For example, when invoke the ``chat`` service, an application may follow the
style of ``ollama``, that means, it tries to access URL like ``/api/chat``, and
it specifies the ``temperature`` parameter in ``options.temperature`` of the
JSON body.

However, the underlying service provider responding this request, may be an
OpenAI like implementation. It expects the URL to be ``/v1/chat/completions``
and the ``temperature`` parameter to be in the root of the JSON body.

.. list-table:: Examples chat service in ollama API Flavor and openai API Flavor
   :header-rows: 1

   * - ``chat`` service
     - **ollama Flavor**
     - **openai Flavor**
   * - **Endpoint**
     - POST /api/chat
     - POST /v1/chat/completions
   * - **Request Schema**
     - See ollama's doc
     - See openai's doc
   * - **Response Schema**
     - See ollama's doc
     - See openai's doc
   * - **Temperature Parameter**
     - options.temperature
     - temperature
   * - **Others**
     - ...
     - ...



``Byze`` calls such style as ``API Flavor`` or ``Flavor`` for short. In above
example, the application calls the ``chat`` with ``ollama Flavor`` but the
service provider provides the service in ``openai Flavor``, neither of them
directly using ``byze Flavor``.

``Byze`` tries to do conversion between different ``Flavor`` so application
doesn't need to worry about the details of the service provider's API. Details in 
:ref:`flavor_conversion`.


