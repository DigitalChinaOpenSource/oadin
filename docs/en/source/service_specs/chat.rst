===================================
Chat Related Byze Services
===================================

chat Service
=====================

.. _`custom_properties_chat`:

Custom Properties of its Service Providers
--------------------------------------------

In addition to the common properties defined in :ref:`Metadata of Byze Service
Provider`, the chat service providers may also have the following properties put
into the ``custom_properties`` field of the service provider metadata.

.. list-table::
   :header-rows: 1

   * - Custom Property
     - Value
     - Description
   * - max_input_tokens
     - integer
     - Width of context window or maximum number of allowed input tokens

Request Schema
--------------------------------------------

.. _`header_chat`:

Header
___________

See :ref:`Common Fields in Header of Request`


.. _`request_chat`:

Request
______________

In addition to these defined in :ref:`Common Fields in Request Body`, the 
service may also have the following fields in its Request JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - messages
     - See :ref:`message_type`
     - required
     - chat messages, may with the conversation history
   * - seed
     - integer
     - optional
     - to help return deterministic results
   * - temperature
     - number between 0 to 2, and default is 1
     - optional
     - Increasing the temperature will make the model answer more creatively.
   * - top_p
     - float
     - optional
     - higher top_p leads to more diverse text, while a lower value (e.g., 0.5)
       produces more focused and conservative text. Default is 0.9.

.. _`response_chat`:

Response Schema
--------------------------------------------

In addition to these defined in :ref:`Common Fields in Response Body`, the 
service may also have the following fields in its Response JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - message
     - See :ref:`message_type`
     - required
     - returned message
   * - finished
     - ``true`` or ``false``
     - required
     - ``true`` for last message in stream mode, otherwise ``false``. It is always
       ``true`` for sync mode
   * - finish_reason 
     - stop, length, function_call, or null
     - required when finished is true
     - | stop for normal ending
       | length for reaching the maximum length
       | function_call for ending due to function call
       | null means not finished yet and the finshed should be false

Examples
--------------

Sending Request

.. code-block:: shell

    curl https://localhost:6688/byze/v0.1/chat/SERVICES/completions \
    -H "Content-Type: application/json" \
    -d '{
        "messages": [
            {
                "role": "system",
                "content": "You are a helpful assistant."
            },
            {
                "role": "user",
                "content": "Hello!"
            }
            ],
            "stream": true,
            "hybrid": true
    }'

Returned Response

.. code-block:: json

    {
        "byze": {
            "non_byze_data_in_response": {
                "prompt_eval_count": 26,
                "... other ollama specific data here ...": "..."
            },
            "received_at": "2024-06-26T19:22:26.123127",
            "responsed_at": "2024-06-26T19:22:28.123127",
            "served_by": "local",
            "served_by_api_flavor": "ollama"
        },
        "message": {
            "role": "assistant",
            "content": "hello "
        },
        "finished": false
    }

    {
        "byze": {
            "non_byze_data_in_response": {
                "prompt_eval_count": 26,
                "... other ollama specific data here ...": "..."
            },
            "received_at": "2024-06-26T19:22:26.123127",
            "responsed_at": "2024-06-26T19:22:28.666127",
            "served_by": "local",
            "served_by_api_flavor": "ollama"
        },
        "message": {
            "role": "assistant",
            "content": "world"
        },
        "finished": true,
        "finish_reason": "stop"
    }




function_call Service
=====================

The function_call is very smilar as the chat service. Many cloud vendors provide
the function_call directly through the same endpoint of chat service, while have
some additional fields in the request body to specify the function call.

Byze specifically create a new end point to seperate it from the more general
chat service. So it clearly tells whether the platform has the service provider
which is capable enough to provide support of function call (even through chat)


Custom Properties of its Service Providers
--------------------------------------------

See :ref:`Custom Properties of chat Service Providers <custom_properties_chat>`

Request Schema
--------------------------------------------

Header
___________

See :ref:`Common Fields in Header of Request <header_chat>`


Request
______________

In addition to these defined in :ref:`Common Fields in Request Body of chat
Service <request_chat>`, the service may also have the following fields in its
Request JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - tools
     - A list of :ref:`Tool Description`
     - optional
     - A list of tools the model may call.
   * - tool_choice
     - none, required, or a JSON object like ``{"type": "function", "function":
       {"name": "tool_name"}}``
     - optional
     - Controls which (if any) tool is called by the model. The object forces the
       model to call that function.

Response Schema
--------------------------------------------

It contains all of these defined in :ref:`response_chat`.

However, the ``message`` in the returned response, may have an additional field
so called ``tool_calls``, which is a list of function invocation (or tool calls)
suggested by the LLM. Application may double check and call them, put the result
in :ref:`Tool Message` and invoke this service again as history to continue the
conversation.

In addition to these defined in :ref:`Common Fields in Response Body`, the 
service may also have the following fields in its Response JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - tool_calls
     - a list of :ref:`Tool Call Description`
     - optional
     - LLM suggested function calls


Examples
------------------------

A concrete example is actually in :ref:`Message Example`. Specifically, it
provides the search function/tool to underlying LLM, which then suggests the
invocation of that function in a response looks like this.

.. code-block:: json

    {
        "byze": {
            "received_at": "2024-06-26T19:22:26.123127",
            "responsed_at": "2024-06-26T19:22:28.123127",
            "served_by": "https://api.openai.com/v1/chat/completions",
            "served_by_api_flavor": "openai"
        },
        "message": {
            "role": "assistant",
            "tool_calls": [{
                "id": "call_BEGxtsoiM96M78Y97RFxPRYk", 
                "type": "function", 
                "function": {"name": "search", "arguments": "{'query':'shirts'}"}
            }]
        },
        "finished": true,
        "finish_reason": "function_call"
    }




text_embed Service
=====================


Custom Properties of its Service Providers
--------------------------------------------

In addition to the common properties defined in :ref:`Metadata of Byze Service
Provider`, the chat service providers may also have the following properties put
into the ``custom_properties`` field of the service provider metadata.

.. list-table::
   :header-rows: 1

   * - Custom Property
     - Value
     - Description
   * - max_input_tokens
     - integer
     - Width of context window or maximum number of allowed input tokens

Request Schema
--------------------------------------------

Header
___________

See :ref:`Common Fields in Header of Request`

Request
______________

In addition to these defined in :ref:`Common Fields in Request Body`, the 
service may also have the following fields in its Request JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - input 
     - string
     - required
     - Input text for the embedding


Response Schema
--------------------------------------------

In addition to these defined in :ref:`Common Fields in Response Body`, the 
service may also have the following fields in its Response JSON body:

.. list-table::
   :header-rows: 1
   :widths: 10 35 10 45

   * - Additional JSON Field
     - Value
     - Required
     - Description
   * - embedding
     - Array of float
     - required
     - returned embedding

Examples
----------------

The returned embedding may look like this

.. code-block:: json

    {
        "byze": {
            "received_at": "2024-06-26T19:22:26.123127",
            "responsed_at": "2024-06-26T19:22:28.123127",
            "served_by": "http://localhost:11434/api/embed",
            "served_by_api_flavor": "openai"
        },
        "embedding": [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
    }
