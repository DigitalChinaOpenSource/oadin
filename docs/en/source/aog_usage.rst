===========================================
How to Use Byze
===========================================

.. include:: global.rst

Byze (AIPC Open Gateway) is a runtime aiming at providing an extremely simple and
easy infrastructure for developers to install local AI services on their
development environments, and to ship their AI applications without the need of
packaging own AI stacks and models.





.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Application Utilizing Byze"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 


     subgraph cluster_aipc {
        label = "AIPC"
        color="#dddddd"
        fillcolor="#eeeeee"

        app_a[label="Application A", fillcolor="#eeeeff"]
        app_b[label="Application B", fillcolor="#eeeeff"]
        app_c[label="Application C", fillcolor="#eeeeff"]

        byze[label="Byze API Layer", fillcolor="#ffffcc"]


        subgraph cluster_service {
            label = "Byze AI Service Providers"
            color = "#333333"
            fillcolor="#ffcccc"

            models[label="AI Models", fillcolor="#eeffcc"]
        }

        {app_a, app_b, app_c} -> byze
        byze -> models[lhead=cluster_service, minlen=2]
     }
     cloud[label="Cloud AI Service Providers", fillcolor="#ffcccc"]
     byze -> cloud[minlen=2 style="dashed"]



   }


As illustrated by the figure above, Byze provides platform-wise AI services so
multiple coexising AI applications don't need to redundantly ship and launch
their own AI stack. This significantly reduces the application size, eliminates
repeated downloading of same AI stack and models for each app, and avoids
competing memory consumption during execution. 

Byze provides the following basic features:

* **One-stop AI service installation**
  
  During development, developers can install AI services locally on their
  development environments by simple commands like ``byze install chat`` or ``byze
  pull-model deepseek-r1-1.5b for chat``. Byze automatically downloads and
  installs the most suitable and optimized AI stacks (e.g. ollama) and models.

  During deployment, developers can ship their AI application without the need
  to package dependent AI stacks and models. Byze will automatically pull the
  required AI stacks and models for the deployed PC when needed. 
  

* **Decouple application and AI service providers with shared service & standard API**

  Byze API Layer offered standandized API for typical AI services like chat,
  embed etc. Developers focus on business logic of their application, without taking 
  too much care about underlying AI service stack. 

  The AI service is provided at platform wise, and shared by multiple
  applicaitons on the same system. This avoid


* **Easy migration by auto API conversion between popular API styles**

  Furthermore, Byze API Layer also provides auto API conversion between popular
  API styles (e.g. OpenAI API) and Byze provided AI service.

  So developers can easily migrate their existing Cloud AI based application to
  Byze based AIPC application.
  
* **Hybrid scheduling between local and cloud AI Service Providers**

  Byze allows developers to install AI services
  locally on their development environments. These services can be accessed
  through the Byze API layer.


Build Byze Command Line Tool
==================================

To build Byze, you need to have `golang <https://go.dev/>`_ installed on your
system.

Then download or clone this project to a directory such as ``/path_to_byze``.

Then run the following commands:

.. code-block:: bash

    cd /path_to_byze
    cd cmd
    go build -o byze

This will generate an executable file named ``byze`` which is the command line of
Byze.

Use Byze Command Line Tool
=================================

You may type ``byze -h`` to see the help information of the command line tool.

Here are some examples

.. code-block:: bash

    # install AI services to local
    # Byze will install neccessary AI stack (e.g. ollama) and models
    byze –install chat
    byze –install audio/text-to-speech
    byze –install embed

    # you may install more models to the service in addition to the default one
    byze –pull-model deepseek-r1-1.5b -for chat

    # install multiple services / models at once
    byze –install –r requirements.txt

    # upgrade Byze
    byze –upgrade

And use these commands to start and stop Byze service

.. code-block:: bash

    byze start

    # start Byze in daemon mode
    byze start -d

    byze stop

Invoke Byze API
=========================

Byze API is a RESTful API. You can invoke the API in similar way as you are invoking
a cloud AI service such as OpenAI. Detailed API specification can be found in
`Byze API Specification <byze_spec>`_.

For example, you may use curl to test the chat service on Windows.

.. code-block:: bash

    curl -X POST http://localhost:16688/byze/v0.2/services/chat  -X POST -H
    "Content-Type: application/json" -d
    "{\"model\":\"glm4\",\"messages\":[{\"role\":\"user\",\"content\":\"why is
    the sky blue?\"}],\"options\":{\"seed\":12345,\"temperature\":0},
    \"hybrid_policy\":\"always_local\", \"stream\":false}" 

Furthermore, if you already have an application using OpenAI API or ollama API
etc., you do NOT need to rewrite the way of invoking Byze to comply with its spec.

Byze is able to convert the API between these popular styles automatically, so you
can easily migrate the app by simply changing the endpoint URL.

For example, if you have an application using the chat completions service of
OpenAI, you may simply replace the endpoint URL from
``https://api.openai.com/v1/chat/completions`` to
``http://localhost:16688/byze/v0.2/api_flavors/openai/v1/chat/completions``.

NOTE that the new URL to invoke Byze is under ``api_flavors/openai`` and rest of
the URL is the same as the original OpenAI API, i.e. ``/v1/chat/completions``.

If you are using ollama API, you may replace the endpoint URL from
``https://localhost:11434/api/chat`` to
``http://localhost:16688/byze/v0.2/api_flavors/ollama/api/chat``. Similarly, it
is under ``api_flavors/ollama`` and rest of the URL is the same as the original 
ollama API, i.e. ``/api/chat``.

Ship your Byze based AI application
==========================================

To ship your AI application, you only need to package your application with a
tiny Byze component so called ``Byze Checker`` which is ``byze.dll`` on Windows.
You don't need to ship AI stack or models.

Using a C/C++/C# application as example, these are the steps to ship a Byze based
AI application.

1. prepare a ``.byze`` file along with your application. The ``.byze`` file is text
manifest file that specifies the AI services and models your application needs.
For example, a ``.byze`` file may look like this:

.. code-block:: json

    {
        "version": "0.2",
        "services": [
            {
                "name": "chat",
                "models": [
                    {
                        "name": "deepseek-r1-1.5b"
                    }
                ]
            },
            {
                "name": "audio/text-to-speech"
            }
        ]
    }


2. Include ``byze.h`` and invoke ``ByzeInit()`` in your ``main()`` function. The
``ByzeInit()`` will:

    * check if Byze is installed on the target PC. If not, it will automatically
      download and install Byze.
    * check if the required AI services and models (as manifested in ``.byze``
      file) are installed. If not, it will automatically download and install
      them.

3. Link the application with the ``byze.dll``.

4. Ship the application with the ``.byze`` file and ``byze.dll`` in the same
   directory as the ``.exe`` file of your application.




TODO:

* specify the .byze file format -- requirements.txt into spec
* describe the detailed init process somewhere