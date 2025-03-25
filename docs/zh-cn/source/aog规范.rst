==========================
 Byze 通用规范
==========================

.. include:: global.rst

版本控制
====================

当前版本 |byze_spec_version|

版本遵循语义版本控制，基本上是：

- 主版本号，当进行不兼容的 API 更改时
- 次要版本号，添加向后兼容功能时
- 补丁版本，修复向后兼容错误时

因此，主版本号是访问 ``Byze APIs`` URL 的一部分


Byze 服务
====================

``Byze Service`` 是一个提供人工智能功能的服务。它可以是一个聊天机器人、文本转语音服务、翻译服务等。每个 ``Byze Service`` 都有它独特的 ``properties`` 和 ``API`` 。

* ``properties`` 描述了服务的特征，例如它所使用的模型、采用的混合策略等。

* ``API`` 描述了服务的接口，例如输入/输出格式、所使用的 HTTP 方法等。

Byze 服务名称和端口
-----------------------------------------

每个 ``Byze Service`` 都有一个独特的名称。该名称用于识别服务，并用于形成访问服务的端点 URL。

 ``Byze Service`` 的典型名称包括 ``chat``, ``audio/text-to-speech``, ``embed`` 等。

``Byze`` 截至目前支持 Restful 风格的 API。API 的 ``Endpoint`` 包括 HTTP 方法和访问服务的 URL。
URL 由主机名和端口（默认为 localhost 和 16688） Byze 前缀、规范版本号以及服务名称组成，即 ``http://localhost:{port_number}/byze/v{spec_version_number}/services/{service_name}``.

例如 ``chat`` 服务的 ``Endpoint`` 是 ``POST http://localhost:16688/byze/v0.2/services/chat`` 。

.. _`Metadata of Byze Service`:

Byze 服务元数据
-----------------------------------

Byze 服务的元数据是一个描述服务的 JSON 对象。它可能因服务而异。然而，在服务的元数据中通常包含一些常见的字段。

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - 字段
     - 值
     - 描述
   * - hybrid_policy
     - ``always_remote``, ``always_local``, ``default``
     - 混合策略的使用
   * - local_service_providers
     - 用于引用： :ref:`Metadata of Byze Service Provider`
     - 默认的本地服务提供商名称
   * - remote_service_providers
     - 用于引用： :ref:`Metadata of Byze Service Provider`
     - 默认的远端服务提供商名称


Byze 服务提供商
=========================

Byze 服务提供商实际上实施了相应的 Byze 服务。

通常，对于一个给定的 Byze 服务，它可能有多个服务提供商。例如，一个 ``/chat Byze`` 服务可能由 ollama 在同一台机器上运行的本地 Byze 服务提供商提供，同时还有一个由 OpenAI 提供的远程 Byze 服务提供商。

每个 Byze 服务提供商都有一个唯一的 ID。该 ID 用于识别服务提供商，并用于引用服务提供商，例如在 Byze 服务的元数据中。参见： :ref:`Metadata of Byze Service`.

.. _`Metadata of Byze Service Provider`:

Byze 服务提供商元数据
------------------------------------

Byze 服务提供商的元数据是一个描述服务提供商的 JSON 对象。以下是大多数服务提供商都有的常见字段。

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - 字段
     - 值
     - 描述
   * - provider_name
     - string
     - 服务提供商的名称/ID
   * - service_name
     - string
     - 当前提供商所提供的服务的名称/ID
   * - service_source
     - ``remote`` or ``local``. Default is ``local``
     - 服务来源
   * - desc
     - string
     - 服务提供商的描述
   * - method
     - string, 例如, ``POST``
     - 调用服务提供商时使用的 HTTP 方法
   * - url
     - string
     - 访问服务提供商的 URL。这与方法共同构成服务提供商的端口
   * - flavor
     - string, 例如, ``byze``, ``openai``, ``ollama``, etc.
     - 服务提供商的 API 版本。如果需要，用于转换 API。参见： :ref:`flavor_conversion`
   * - properties
     -  ``{"max_input_tokens": 2048,"supported_response_mode":["stream","sync"]}``
     - 当前服务特有的属性，包含 最大上下文、是否支持流式输出等
   * - auth_type
     - ``none`` or ``apikey`` or ``token``. Default is ``none``
     - 鉴权方式
   * - auth_key
     - string
     - 鉴权信息， 根据auth_type的不同存储不同的鉴权信息，例如 none: {} or apikey: "{'apikey':'xxx'}" or token: "{'ak':'xxx','sk':'xxx'}"
   * - extra_headers
     - JSON object
     - 发送请求时需要在头部添加的其他内容吗。例如，与授权相关的头部
   * - extra_json_body
     - JSON object
     - 其他需要添加到该云服务请求体中的内容
   * - status
     - ``0`` or ``1``. Default is ``0``
     - 当前服务提供商提供的服务状态：0-不可用 1-可用


通用数据类型
=========================

在 ``Byze Service`` 的 Restful API 中，有一些常用的数据类型。例如，如何表示文本图像等。我们在这里描述这些通用类型。


.. _`content_type`:

内容
-----------

内容可以出现在许多请求或响应的 JSON 体中。内容可以是各种类型。

文本内容
_____________

文本内容可以是字符串

.. code-block:: json

    "content": "a plan string content"


或指定类型的对象

.. code-block:: json

    {
        "type": "text",
        "text": "the text content here"
    }


它可以通过添加更多信息变得更加复杂

.. code-block:: json

    {
        "type": "text",
        "text": {
            "value": "the text content",
            "annotations": ["tag1"]
        }
    }

图像内容
_____________

图像内容可以是 base64 编码的字符串，例如，下表中的两个 base64 编码的图像。

.. code-block:: json

    "images": ["iVBORw0KGgoAAAAN...", "YAAADBPx+VAA..."]


或者它可能是一个带有 URL 的对象

.. code-block:: json

    {
        "type": "image_url",
        "image_url": {
            "url": "http://a.com/b.jpg"
        } 
    }


或 base64 编码字符串

.. code-block:: json

    {
        "type": "image":
        "image": "iVBORw0KGgoAAAAN..."
    }


OpenAI 还提供了 ```image_file``` ，它也在 Byze 的未来路线图中。


Array Content
_____________

数组内容只是一个列表，其中包含上面定义的内容。

.. _`message_type`:

消息
---------------------------

存在几种消息类型，每种类型都有自己的字段。迄今为止，我们仅支持系统、用户、助手、工具消息类型。

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 15 55

   * - 字段
     - 值
     - 是否必需
     - 描述
   * - role
     - ``system``, ``user``, ``assistant``, or ``tool``
     - 必需
     -  
   * - content
     - 见 :ref:`content_type`
     - 必需
     - 
   * - tool_call_id
     - a string id
     - 可选，如果 role 是 ``tool`` 则必需
     - 工具调用，该消息正在响应。这与 OpenAI 在聊天/完成功能中的函数调用能力相似 ``chat/completions``


.. _`Message Example`:

消息实例
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

工具
-----------------

工具消息是一种通常包含函数调用执行结果的消息，作为对话或历史记录的一部分发送回大模型。

.. _`Tool Message`: 

``tool_call_id`` 用来指示这条消息响应的是哪个调用。 ``content`` 通常是所需工具调用LLM的执行结果。 详见上文 :ref:`message_type` 和 :ref:`Message
Example`.

.. _`Tool Description`:

工具描述
___________________________________

这描述了一个可用的函数，并将其提供给LLM。据此，LLM可以决定是否建议调用此函数，以及使用哪种类型的参数。因此，它不仅应该有关于函数的名称和描述，还应该有关于其参数的架构。

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 15 15 40

   * - 字段
     - 下级字段
     - 值
     - 是否必需
     - 描述
   * - type 
     - 
     - "function"
     - 必需
     - 指定这是一个函数类型
   * - function
     - 
     - 
     - 
     -
   * - 
     - name
     - string 
     - 必需
     -
   * - 
     - description
     - string
     - 可选
     - 函数功能的描述，用于模型选择何时以及如何调用该函数。
   * -
     - parameters
     - JSON object
     - 可选
     - 函数接受的参数，描述为一个 JSON 模式对象 `JSON
       Schema <https://json-schema.org/understanding-json-schema>`_ 

这里是一个例子，同时你可以在 :ref:`Message Example` 中找到另一个例子

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

工具调用描述
_______________________

这描述了调用函数的请求。它通常作为消息的一部分由 ``tool_calls`` 中的 LLM 返回，前提是提供了可用的： :ref:`Tool
Description`.

.. list-table:: 
   :header-rows: 1
   :widths: 10 10 10 10 60

   * - 字段
     - 下级字段
     - 值
     - 是否必需
     - 描述
   * - id 
     - 
     - string
     - 必需
     - 工具ID
   * - type
     - 
     - function
     - 必需
     - 
   * - function
     - 
     - 
     - 
     -
   * - 
     - name
     - string 
     - 必需
     - 要调用的函数名称
   * - 
     - arguments
     - string
     - 可选
     - 函数调用参数，由模型以 JSON 格式生成。请注意，模型并不总是生成有效的 JSON，可能会生成由您的函数模式未定义的参数。在调用函数之前，请在您的代码中验证这些参数。

这是从上述： :ref:`Message Example` 中提取的示例。

.. code-block:: json

    {
        "id": "call_BEGxtsoiM96M78Y97RFxPRYk", 
        "type": "function", 
        "function": {
            "name": "search", 
            "arguments": "{'query':'shirts'}"
        }
    }

字段：服务头部字段和 JSON 正文字段
======================================================

服务通过 HTTP 调用，Byze 服务要求请求和响应都使用 ``JSON`` 格式。

必填字段和可选字段
------------------------------

请求/响应头字段以及请求/响应的 JSON 体字段，要么是必需的，要么是可选的。

必填和可选字段
___________________________________________

必填字段应在调用时提供，即发送请求。它们通常是所有潜在服务提供商实际运营服务所需的最基本信息集。

可选字段可由应用程序提供，但无法保证底层实际服务提供商支持它们。

**NOTE** 请注意，如果基础实际服务提供商不支持可选字段，则在 Byze 转换过程中将从请求中删除该字段。
这可能会导致某些信息丢失，但可避免意外将参数发送给服务提供商。此行为在未来规范中可能有所更改。更多详情请参阅： :ref:`flavor_conversion`.


.. list-table:: 
   :header-rows: 1
   :widths: 20 30 50

   * - 请求字段类型
     - 调用是否需要提供
     - 底层服务提供商是否接受
   * - 必需
     - | 是。
       | 必须提供，除非该字段有默认值
     - | 是。
       | 实际服务提供商将接收这些字段并使用它们
   * - 可选
     - | 否。
       | 调用者可自行决定是否提供这些字段
     - | 视情况而定。
       | 只有实际服务提供商支持的字段会被传递给它。其他字段将被过滤并忽略。

必填和可选字段
____________________________________________

响应中定义的必填字段始终由 Byze 提供。

然而，返回响应中是否会出现可选字段，取决于处理相应请求的实际服务提供商。

与请求不同，Byze 中未定义但由服务提供商返回的附加字段将仍然保留在那里，并直接发送到应用程序而不会被过滤掉。

.. list-table:: 
   :header-rows: 1
   :widths: 20 40 40

   * - 响应字段类型
     - 是否出现在底层服务提供商返回的响应中
     - 是否出现在 Byze 返回的最终响应中
   * - 必需
     - | 未定。 
       | 底层服务提供商的响应可能包含相关信息，但未放入此字段。
     - | 会。 
       | Byze 将转换并确保该字段被展示，并且该字段可安全访问
   * - 可选
     - 未定
     - | 未定。
       | Byze 最终返回的响应仅包含由实际服务提供商提供的部分

.. _`Common Fields of Byze Services`:

Byze 服务通用字段
==================================

尽管每个服务在其请求/响应的头部和 JSON 主体中可能有自己的特定字段。一些字段相当常见，出现在大多数服务请求/响应中。

.. _`Common Fields in Header of Request`:

请求头中的常见字段
--------------------------------------

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 10 65

   * - 字段
     - 值
     - 是否必需
     - 描述
   * - Content-Type
     - application/json
     - 必需
     - 除非特别说明，大多数 API 端点仅接受 JSON 格式

大多数 API 以 JSON 作为输入数据，因此头部包含 ``Content-Type`` 。 然而，对于二进制输入，如图像或音频文件等，API 可能会使用表单数据 (例如， ``multipart/form-data`` 在 ``Content-Type`` 中)

**NOTE** 请注意，与调用云服务不同， ``Authorization`` 头不是始终必需的，因为它可能本地调用 Byze 服务。然而，Byze 可以根据配置或应用程序调用的信息稍后添加此类头。

.. _`Common Fields in Request Body`:

请求体中的常见字段
---------------------------------

.. list-table:: 
   :header-rows: 1
   :widths: 15 15 10 60

   * - JSON 字段
     - 值
     - 是否必需
     - 描述
   * - model
     - string or JSON object
     - optional
     - 要用于此服务的模型。如未提供，将使用默认模型。如果服务不支持选择模型，则此字段将被忽略。参见 
       :ref:`Metadata of Byze Service Provider` 中的 ``allow_to_select_model`` 。
       某些服务可能有多个模型，因此应用程序可能传递一个 JSON 对象来指定服务内部使用的每个模型的模型名称。
   * - stream
     - ``true`` 或 ``false``. 默认为 ``false``
     - 必需
     - 是否使用流模式。如未提供，将使用默认模式。参见： :ref:`Metadata of Byze Service Provider` ``supported_response_mode``
   * - hybrid_policy
     - ``always_remote``, ``always_local``, ``default``
     - 可选
     - 混合策略使用。如未提供，将使用 ``default`` 策略。参见 :ref:`Metadata of Byze Service` 中的 ``hybrid_policy``
   * - remote_service_provider
     - string, 或 JSON ，如 :ref:`Metadata of Byze Service Provider`
     - 可选
     - 默认情况下，如果 Byze 决定使用远程服务提供商来处理此调用（基于 hybrid_policy 等），它将使用平台定义的默认远程服务提供商为此服务提供服务。
       然而，应用程序可以通过使用此字段在请求中覆盖此特定设置，以确保当 Byze 决定使用远程服务提供商时，使用该远程服务提供商。
       该字段需要是字符串，表示预定义服务提供商的 ID。或者，它也可以是一个 JSON 对象，表示 Byze 服务提供商的： :ref:`Metadata of Byze Service Provider`
   * - keep_alive
     - timespan, e.g. ``5m`` as the default
     - 可选
     - 底层服务（例如 ``ollama`` ）控制本地模型在内存中停留时间的提示


.. _`Common Fields in Response Body`:

响应体中的常见字段
---------------------------------

在许多情况下，返回的响应是一个 JSON 对象。除了必要时进行转换 (见 :ref:`flavor_conversion`)。

.. list-table:: 
   :header-rows: 1
   :widths: 50 50
   
   * - JSON 字段
     - 描述
   * - model
     - 实际调用的模型。这是可选的，并不总是提供。
