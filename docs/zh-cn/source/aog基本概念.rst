====================================
Byze 关键概念
====================================

.. include:: global.rst

在这一节中，我们将介绍 ``Byze`` 的关键概念和术语.

Byze 服务和 Byze 服务提供商
=========================================

在 ``Byze`` 中, 典型的通用 AI 服务例如 ``chat`` 和 ``text-to-image``
等被定义为 ``Byze Service``，每个服务都有其标准化的API.

目前， ``Byze`` 支持基于HTTP协议的RESTFUL API。 所以这里的API指的是接口（HTTP verb 和 URL），以及通过该接口进行对应服务的请求和响应的格式。

另一方面 ``Byze Service Provider`` 是实际提供服务的实体，它实现了 ``Byze Service`` 的API。

这与面向对象编程（OOP, Object Oriented Programming）中的抽象接口和具体类概念相似，即 ``Byze Service`` 是一种抽象接口，而 ``Byze Service Provider`` 是实现该接口的具体类。


实际操作上，一个 ``Byze Service`` 可以拥有多个 ``Byze Service Providers`` 。一个 ``Byze Service Provider`` 也可以提供多个 ``Byze Services`` 。


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

如上图所示，远程 *OpenAI Chat* 和本地 *Ollama Chat* 均为 ``Byze Service Providers`` 服务的 ``chat`` 。同时，*rag/summarize* 服务实际上依赖于与 ``chat`` 相同的提供商。



Byze 服务详情
================================

一个 ``Byze Service`` 包含以下定义属性：

Name
    服务的名称。 这是 ``Byze Service`` 的字符串类型的ID。因此，每个服务的名称都是唯一的。例如
    ``chat``, ``embed``, ``rag/query``, ``text-to-image``,
    ``audio/text-to-speech`` 等。

HybridPolicy
    服务的调度策略。在服务层面可以有多种调度策略，我们期望在单个服务拥有多个不同服务提供商时，
    可以通过一定的策略来自动选择和切换不同的服务商来提供 AI 服务，这部分目前暂仅支持以下三种策略，
    在未来我们期望在调度策略这里加入更多的可能性。
    例如 ``default``, ``always_remote``, ``always_local`` 等。

RemoteProvider
    默认的远端服务提供商的名称，由于我们单个服务可能会有多个不同的远端服务提供商，所以我们需要用户指定
    一个默认的远端服务提供商。当然，在我们初次安装远端服务时，我们会自动将当前服务提供商设置为默认远端服务提供商。

LocalProvider
    默认的本地服务提供商的名称，由于我们在未来会支持多种本地模型引擎（ ``openvino``、 ``ollama`` ）等，
    和 ``RemoteProvider`` 相同，我们同样需要指定一个默认的本地服务提供商。当然，在我们初次安装本地服务时，
    我们会自动将当前服务提供商设置为默认本地服务提供商。


Byze 服务提供商详情
========================================================

ProviderName
    服务提供商的名称。

ServiceName
    服务名称。表明当前服务提供商提供了哪种服务。

ServiceSource
    服务来源。表明服务来源于何处，这里指的是： ``local`` 或 ``remote``

Flavor
    服务提供商的风格类别。用于区分服务提供商的类别，例如 ``openai``、 ``ollama``、
    ``deepseek``、 ``tencent`` 等。

Method
    服务的调用方式，即 HTTP verb （GET/POST）

URL
    服务的实际调用地址，例： ``http://localhost:11434/api/chat``

AuthType
    服务的鉴权方式。由于远端 AI 服务基本都需要鉴权才能访问，所以这里我们提供了鉴权支持，例如：
    ``none``, ``apikey``, ``token``。

AuthKey
    服务的鉴权信息。根据 ``AuthType`` 提供不同的鉴权信息。

.. _byze_service_provider_properties:

Properties
    每个 ``Byze Service`` 也有一组属性。这些属性用于描述实现服务的实际服务提供商，因此其值因服务提供商而异。

    例如， ``chat`` 服务有 ``max_input_tokens`` 属性，该属性告知上层应用程序可以向底层服务提供商发送的最大token数。对不同的服务提供商，该属性的值也不同。



Byze 应用的 API 风格
========================================================

在 :doc:`byze规范` 中， ``Byze`` 定义了 ``Byze Services`` 的规范，包括其名称、端口、请求模式、响应模式等。

然而实际上应用开发者和服务提供商很可能并未完全遵守 ``Byze`` 的规范，正如文档 :ref:`compatibility_issue` 所说的。

例如，当调用 ``chat`` 服务时，应用程序可能遵循 ``ollama`` 的风格，即它会访问类似 ``/api/chat`` 的 ``URL``，并在 JSON 体的 ``options.temperature`` 中指定 ``temperature`` 参数。

然而，响应此请求的底层服务提供商可能是一个类似于 OpenAI 的实现。它期望 URL 为 ``/v1/chat/completions`` ，并且 ``temperature`` 参数位于 JSON 主体的根部。

.. list-table:: ollama 风格的 API 和 openai 风格的 API
   :header-rows: 1

   * - ``chat`` 服务
     - **ollama 风格**
     - **openai 风格**
   * - **Endpoint**
     - POST /api/chat
     - POST /v1/chat/completions
   * - **Request Schema**
     - 见ollama的文档
     - 见openai的文档
   * - **Response Schema**
     - 见ollama的文档
     - 见openai的文档
   * - **Temperature Parameter**
     - options.temperature
     - temperature
   * - **其他**
     - ...
     - ...



``Byze`` 将这种风格简称为 ``API Flavor`` 或 ``Flavor`` 。在上例中，应用程序调用 ``chat`` 使用 ``ollama Flavor`` ，
但服务提供商在 ``openai Flavor`` 提供服务，他们都没有直接使用 ``byze Flavor`` 。

Byze 会在不同 Flavor 之间进行转换，因此应用程序无需担心服务提供商 API 的细节。详细信息请参阅：
:ref:`flavor_conversion`.
