========================================
Problem Statements targetted by Byze
========================================

Bad Pattern of Ship Everything
====================================================

Today, most applications require integration from the ground up when
incorporating AI-related features such as "Help Me Write." This often results in
shipping everything within the application as a monolithic distribution package.
For instance, the application package might need to include the LLM model, the
Ollama inference and serving framework, along with the rest of the
application.


.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     label = "Applications of Shipping Everything"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 

     subgraph cluster_aipc {
       label = "AIPC"
       color="#dddddd"
       fillcolor="#eeeeee"

       subgraph cluster_application_a {
            label = "Application A"
            color="#333333"
            fillcolor="#eeeeff"

            subgraph cluster_service_a {
                label = "AI Services with X APIs"
                fillcolor="#ffcccc"

                models_a[label="AI Models", fillcolor="#eeffcc"]
            }
       }

       subgraph cluster_application_b {
            label = "Application B"
            color="#333333"
            fillcolor="#eeeeff"

            subgraph cluster_service_b {
                label = "AI Services with Y APIs"
                fillcolor="#ffcccc"

                models_b[label="AI Models", fillcolor="#eeffcc"]
            }
       }

       subgraph cluster_application_c {
            label = "Application C"
            color="#333333"
            fillcolor="#eeeeff"

            subgraph cluster_service_c {
                label = "AI Services with Z APIs"
                fillcolor="#ffcccc"

                models_c[label="AI Models", fillcolor="#eeffcc"]
            }
       }

     }
   }


This creates a situation on AIPC, as shown in the diagram above. On a PC, each
AI-related application includes its own models and associated AI frameworks,
following what we call the "app-ships-everything" pattern. This approach faces
several challenges:

- The models and AI frameworks bundled with the app are often tuned or optimized
  only for a specific platform (e.g., an AIPC from a particular vendor), while
  not benefiting other platforms.

- Each AI application becomes significantly larger, leading to longer download
  and update time. Multiple applications can also collectively consume
  substantial disk space on the PC.

- Since AIPC has limited computational resources (CPU, GPU, NPU) and memory,
  running multiple AI applications simultaneously is challenging. Each
  application must host its own models, creating resource competition among
  them.


Decoupling AI Application from General AI Services 
====================================================

In practice, most applications, with the exception of a few highly specialized
AI-driven ones, primarily depend on a few general-purpose AI services. For
example, multiple applications within a single AIPC may share the common need
for text editing features. By providing a centralized, optimized 'Help Me Write'
AI service, the platform can significantly improve efficiency and reduce
redundancy. This shared service empowers applications to focus on their core
functionalities, avoiding the overhead of developing and maintaining their own
text editing capabilities. 



.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Application Decoupled from AI Services"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 

     subgraph cluster_aipc {
        label = "AIPC"
        color="#dddddd"
        fillcolor="#eeeeee"

        app_a[label="Application A", fillcolor="#eeeeff"]
        app_b[label="Application B", fillcolor="#eeeeff"]
        app_c[label="Application C", fillcolor="#eeeeff"]


        subgraph cluster_service {
            label = "AI Services @ Platform"
            color = "#333333"
            fillcolor="#ffcccc"


            models[label="AI Models", fillcolor="#eeffcc"]
        }

        {rank=same; app_a app_b app_c}

        app_a -> models[lhead=cluster_service]
        app_b -> models[lhead=cluster_service, minlen=2]
        app_c -> models[lhead=cluster_service]

     }
   }

This new pattern separates AI applications from general AI services.
Applications can concentrate on their core functionalities, calling upon
platform-provided AI services as required. This approach addresses the
limitations of the "ship everything" pattern.


Byze Unified Architecture and API for AI Applications
==========================================================

While the emerging trend of decoupling AI applications from general AI services
offers a solution to the limitations of monolithic AI deployments, it presents a
new set of engineering challenges:

- **Compatibility** - AI applications need to invoke AI services that are not
  shipped by their own, which may have varying APIs, requiring compatibility
  adjustments for seamless deployment across different platforms.

- **Availability** - The application need to be designed to be deployable on any
  PC, even in environments lacking locally installed AI services. In instances
  where local services are unavailable or experiencing high demand, the
  application should seamlessly transition to utilizing remote AI services, such
  as those provided by cloud or edge computing platforms.

``Byze`` (**A**\IPC **O**\pen **G**\ateway) proposes an unified architecture and API
layer to tackle these two problems. 

- ``Byze`` defines a unified ``Byze API`` specification for most general AI services.

- Architecture level, the applications now send their AI requests to ``Byze API
  Layer``. The ``Byze API Layer`` will dispatch the requests to the suitable AI
  services.



.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Application Utilizing Byze API Layer"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 


     subgraph cluster_aipc_1 {
        label = "an AIPC"
        color="#dddddd"
        fillcolor="#eeeeee"

        app_a1[label="Application A", fillcolor="#eeeeff"]
        app_b1[label="Application B", fillcolor="#eeeeff"]

        byze_1[label="Byze API Layer", fillcolor="#ffffcc"]


        subgraph cluster_service_1 {
            label = "AI Services from X"
            color = "#333333"
            fillcolor="#ffcccc"

            models_1[label="AI Models", fillcolor="#eeffcc"]
        }

        {app_a1, app_b1} -> byze_1
        byze_1 -> models_1[lhead=cluster_service_1, minlen=2]
     }
     cloud_m[label="Cloud AI Service from M", fillcolor="#ffcccc"]
     byze_1 -> cloud_m[minlen=2 style="dashed"]

     subgraph cluster_aipc_2 {
        label = "another AIPC"
        color="#dddddd"
        fillcolor="#eeeeee"

        app_a2[label="Application A", fillcolor="#eeeeff"]
        app_b2[label="Application B", fillcolor="#eeeeff"]

        byze_2[label="Byze API Layer", fillcolor="#ffffcc"]


        subgraph cluster_service_2 {
            label = "AI Services from Y"
            color = "#333333"
            fillcolor="#ffcccc"

            models_2[label="AI Models", fillcolor="#eeffcc"]
        }

        {app_a2, app_b2} -> byze_2
        byze_2 -> models_2[lhead=cluster_service_2, minlen=2]
     }
     cloud_n[label="Cloud AI Service from N", fillcolor="#ffcccc"]
     byze_2 -> cloud_n[minlen=2 style="dashed"]

   }

.. _compatibility_issue:

Addressing Compatibility Issues
---------------------------------------

Ideally, applications would interact with AI services exclusively through the
``Byze API``. This would ensure compatibility and enable applications to run on any
platform, regardless of the underlying AI service provider. However, in reality,
some applications or service providers may not fully adhere to the ``Byze API``.

To mitigate this, the ``Byze API Layer`` will attempt to bridge these compatibility
gaps by performing necessary conversions. This is feasible because many APIs,
even if not identical, often share similarities with OpenAI's API. By leveraging
these similarities, the ``Byze API Layer`` can facilitate the interaction between
applications and diverse AI services, even when their APIs are not fully
compatible.


.. _availability_issue:

Addressing Availability Issues
--------------------------------------

The ``Byze API Layer`` acts as a dynamic intermediary, intelligently routing
requests from applications to the most suitable AI service provider. By
monitoring system resources like CPU, GPU, and NPU utilization, it can determine
when local services are overloaded. In such cases, the layer can seamlessly
redirect requests to cloud-based AI service providers, optimizing performance
and ensuring uninterrupted service. This flexibility is made possible by the
``Byze API Layer``'s ability to handle potential API incompatibilities through
conversion and matching, even if the API of cloud-based AI service is slightly
different vs. what application uses.

