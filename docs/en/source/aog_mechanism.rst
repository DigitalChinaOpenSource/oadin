===============================
Oadin Key Mechanisms
===============================


In this section, we will introduce two important mechanisms of ``Oadin`` for two
critical challenges we mentioned earlier: :ref:`compatibility_issue` and
:ref:`availability_issue`.

In addition, we will discuss about how to match the models.


.. _flavor_conversion:

Conversion of API Flavors for Compatibility
===============================================================

``Oadin`` will convert the requests and responses if the ``API Flavor`` of
Application can be different from the ``API Flavor`` of the underlying ``Service
Provider``. 

There are several possible scenarios as listed below. In each scenario,
different conversion will be done by ``Oadin``. The most complicated one is
scenario C, where neither the application nor the service provider uses the
``Oadin Flavor``. In this case, ``Oadin`` will convert the request from the app's
flavor to the ``Oadin Flavor`` first, then convert it to the service provider's
flavor. The response will be converted in the reverse order.


.. list-table:: Conversion of API Flavors for Compatibility
   :header-rows: 1
   :widths: 10 10 10 100

   * - Situation
     - App's Flavor
     - Service Provider's Flavor
     - Conversion done by Oadin
   * - A
     - X
     - X
     - None
   * - B
     - Oadin
     - Oadin
     - None
   * - C
     - X
     - Y
     - | Request: X -> Oadin then Oadin -> Y
       | Response: Y -> Oadin then Oadin -> X
   * - D
     - Oadin
     - Y
     - | Request: Oadin -> Y
       | Response: Y -> Oadin
   * - E
     - X
     - Oadin
     - | Request: X -> Oadin
       | Response: Oadin -> X




.. graphviz::
    :align: center
    
    digraph G {
        rankdir=TB
        compound=true
        label = "Situations of API Flavors"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 10 ]


        subgraph cluster_a {
            label = "A"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_a[label="App", fillcolor="#eeeeff"]
            oadin_a[label="Oadin API Layer", fillcolor="#ffffcc"]
            sp_a[label="Service Provider", fillcolor="#ffcccc"]

            app_a -> oadin_a [label=" X", dir=both]
            oadin_a -> sp_a [label=" X", dir=both]

        }


        subgraph cluster_b {
            label = "B"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_b[label="App", fillcolor="#eeeeff"]
            oadin_b[label="Oadin API Layer", fillcolor="#ffffcc"]
            sp_b[label="Service Provider", fillcolor="#ffcccc"]

            app_b -> oadin_b [label=" Oadin", dir=both]
            oadin_b -> sp_b [label=" Oadin", dir=both]
        }

        subgraph cluster_c {
            label = "C"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_c[label="App", fillcolor="#eeeeff"]
            oadin_c[label="Oadin API Layer", fillcolor="#ffffcc"]
            sp_c[label="Service Provider", fillcolor="#ffcccc"]

            app_c -> oadin_c [label=" X", dir=both]
            oadin_c -> sp_c [label=" Y", dir=both]

        }


        subgraph cluster_d {
            label = "D"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_d[label="App", fillcolor="#eeeeff"]
            oadin_d[label="Oadin API Layer", fillcolor="#ffffcc"]
            sp_d[label="Service Provider", fillcolor="#ffcccc"]

            app_d -> oadin_d [label=" Oadin", dir=both]
            oadin_d -> sp_d [label=" Y", dir=both]
        }

        subgraph cluster_e {
            label = "E"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_e[label="App", fillcolor="#eeeeff"]
            oadin_e[label="Oadin API Layer", fillcolor="#ffffcc"]
            sp_e[label="Service Provider", fillcolor="#ffcccc"]

            app_e -> oadin_e [label=" X", dir=both]
            oadin_e -> sp_e [label=" Oadin", dir=both]
        }

    }



A more detailed flow is illustrated here, for conversion of requests and
responses respectively.


.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Request Body in Oadin API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="Oadin \nReceives \nApp's \nRequest"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_oadin [label="App's Flavor \n==\nOadin ?", shape=diamond]
        is_sp_oadin [label="Flavor of \nService\nProvider\n==\nOadin ?", shape=diamond]
        to_oadin [label="convert\nRequest\nto\nOadin\nFlavor"]
        from_oadin [label="convert to\nFlavor of\nService\nProvider"]
        invoke [label="Invoke\nService\nProvider\nwith its\nFlavor"]

        receive->is_same_flavor
        is_same_flavor->invoke [label="Yes"]
        is_same_flavor->is_app_oadin [label="No"]
        is_app_oadin -> is_sp_oadin [label="Yes"]
        is_app_oadin -> to_oadin [label="No"]
        to_oadin -> is_sp_oadin
        is_sp_oadin -> invoke [label="Yes"]
        is_sp_oadin -> from_oadin [label="No"]
        from_oadin -> invoke

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, invoke
        }

        subgraph r2 {
            rank="same"
            is_app_oadin, to_oadin, is_sp_oadin, from_oadin
        }
    }



.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Response Body in Oadin API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="Oadin \nReceives \nResponse\nfrom\nService\nProvider"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_oadin [label="App's Flavor\n==\nOadin ?", shape=diamond]
        is_sp_oadin [label="Flavor of \nService\nProvider\n==\nOadin ?", shape=diamond]
        to_oadin [label="convert\nResponse\nto\nOadin\nFlavor"]
        from_oadin [label="convert\nto\nApp's\nFlavor"]
        send [label="Send\nResponse\nin App's\nFlavor\nto App"]

        receive->is_same_flavor
        is_same_flavor->send [label="Yes"]
        is_same_flavor->is_sp_oadin [label="No"]
        is_sp_oadin -> is_app_oadin [label="Yes"]
        is_sp_oadin -> to_oadin [label="No"]
        to_oadin -> is_app_oadin
        is_app_oadin -> send [label="Yes"]
        is_app_oadin -> from_oadin [label="No"]
        from_oadin -> send

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, send
        }

        subgraph r2 {
            rank="same"
            is_app_oadin, to_oadin, is_sp_oadin, from_oadin
        }
    }



.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Request Body in Service Provider"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

    }


.. _hybrid_scheduling:

Hybrid Scheduling for Availability
========================================================

``Oadin`` provides hybrid scheduling, i.e. when needed, it will dispatch
application's request (with necessary conversion) to a remote alternative ``Oadin
Service Provider`` (usually a cloud service) instead of local. This is very
helpful when local AIPC is busy, or the desired service is not provided by
current PC, or the user wants to use VIP service at cloud etc.

``Oadin`` makes such dispatch decision by following the specified ``hybrid
policy``. The AIPC with ``Oadin`` installed has a system-wide configuration (See
:doc:`/oadin_platform_config`) which specifies all of the available ``Oadin
Service`` and their corresponding local and remote ``Oadin Service Providers``,
along with the default ``hybrid policy`` to switch between these providers. 

Furthermore, the application can also overwrite the default ``hybrid policy``
defined by the platform config. For example, the application may force to use
the cloud service for a particular request, it can then add ``hybrid_policy:
always_remote`` in the JSON body of request to send.



.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Hybrid Scheduling"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 

     app[label="Application", fillcolor="#eeeeff"]
     oadin[label="Oadin to Dispatch - based on Hybrid Policy", fillcolor="#ffffcc"]
     local[label="Local Oadin Service Provider", fillcolor="#ffcccc"]
     cloud[label="Remote Oadin Service Provider", fillcolor="#ffcccc"]

     app -> oadin

     oadin -> local[style="dashed"]
     oadin -> cloud[style="dashed"]

   }




.. _match_models:

Match Models
========================================================

In a lot of situations, the application may want to specify the preferred model
to use, but the underlying ``Oadin Service Provider`` either doesn't provide the
model, or it provides the model but the name is slightly different.

Currently ``Oadin`` provides a simple mechanism which tries to pick the model from
the service provider which best matches the required model by application. This 
is up to change or evolve in the future.

First, when defines the available ``Oadin Service Provider``, the
:doc:`/oadin_platform_config` can also list the available models for each service
provider, as part of its :ref:`Property of Oadin Service Provider
<oadin_service_provider_properties>`.

Then, the application can specify the model name in the request, for example,
``model: xx-7B`` in its JSON body of the request. ``Oadin`` will do a fuzz match
between this expected model and the available models of the service provider,
and ask to use the most similar one. 

