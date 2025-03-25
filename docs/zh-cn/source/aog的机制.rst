===============================
Byze 关键机制
===============================


在这一节中，我们将介绍针对我们之前提到的两个关键挑战：:ref:`compatibility_issue` 和
:ref:`availability_issue`，这是 ``Byze`` 的两个重要机制。

此外，我们还将讨论如何匹配模型。


.. _flavor_conversion:

转换 API 风格以实现兼容性
===============================================================

如果应用程序的 ``API Flavor`` 与底层 ``Service Provider`` 的 ``API Flavor`` 不同， ``Byze`` 将转换请求和响应。

以下列出几种可能的场景。在每个场景中，将由 ``Byze`` 执行不同风格的转换。其中最复杂的是场景 C，在这种情况下，
既不使用应用也不使用服务提供商的 ``Byze Flavor`` 。在这种情况下， ``Byze`` 将首先将应用的风格转换为 ``Byze Flavor`` ，
然后将其转换为服务提供商的风格，响应将按相反的顺序转换。


.. list-table:: API 风格转换以实现兼容性
   :header-rows: 1
   :widths: 10 10 10 100

   * - 情况
     - 应用的风格
     - 服务提供商的风格
     - Byze 会做的转换
   * - A
     - X
     - X
     - None
   * - B
     - Byze
     - Byze
     - None
   * - C
     - X
     - Y
     - | 请求: X -> Byze then Byze -> Y
       | 响应: Y -> Byze then Byze -> X
   * - D
     - Byze
     - Y
     - | 请求: Byze -> Y
       | 响应: Y -> Byze
   * - E
     - X
     - Byze
     - | 请求: X -> Byze
       | 响应: Byze -> X




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
            byze_a[label="Byze API Layer", fillcolor="#ffffcc"]
            sp_a[label="Service Provider", fillcolor="#ffcccc"]

            app_a -> byze_a [label=" X", dir=both]
            byze_a -> sp_a [label=" X", dir=both]

        }


        subgraph cluster_b {
            label = "B"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_b[label="App", fillcolor="#eeeeff"]
            byze_b[label="Byze API Layer", fillcolor="#ffffcc"]
            sp_b[label="Service Provider", fillcolor="#ffcccc"]

            app_b -> byze_b [label=" Byze", dir=both]
            byze_b -> sp_b [label=" Byze", dir=both]
        }

        subgraph cluster_c {
            label = "C"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_c[label="App", fillcolor="#eeeeff"]
            byze_c[label="Byze API Layer", fillcolor="#ffffcc"]
            sp_c[label="Service Provider", fillcolor="#ffcccc"]

            app_c -> byze_c [label=" X", dir=both]
            byze_c -> sp_c [label=" Y", dir=both]

        }


        subgraph cluster_d {
            label = "D"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_d[label="App", fillcolor="#eeeeff"]
            byze_d[label="Byze API Layer", fillcolor="#ffffcc"]
            sp_d[label="Service Provider", fillcolor="#ffcccc"]

            app_d -> byze_d [label=" Byze", dir=both]
            byze_d -> sp_d [label=" Y", dir=both]
        }

        subgraph cluster_e {
            label = "E"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_e[label="App", fillcolor="#eeeeff"]
            byze_e[label="Byze API Layer", fillcolor="#ffffcc"]
            sp_e[label="Service Provider", fillcolor="#ffcccc"]

            app_e -> byze_e [label=" X", dir=both]
            byze_e -> sp_e [label=" Byze", dir=both]
        }

    }



更详细的流程图在此展示，分别用于请求和响应的转换。


.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Request Body in Byze API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="Byze \nReceives \nApp's \nRequest"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_byze [label="App's Flavor \n==\nByze ?", shape=diamond]
        is_sp_byze [label="Flavor of \nService\nProvider\n==\nByze ?", shape=diamond]
        to_byze [label="convert\nRequest\nto\nByze\nFlavor"]
        from_byze [label="convert to\nFlavor of\nService\nProvider"]
        invoke [label="Invoke\nService\nProvider\nwith its\nFlavor"]

        receive->is_same_flavor
        is_same_flavor->invoke [label="Yes"]
        is_same_flavor->is_app_byze [label="No"]
        is_app_byze -> is_sp_byze [label="Yes"]
        is_app_byze -> to_byze [label="No"]
        to_byze -> is_sp_byze
        is_sp_byze -> invoke [label="Yes"]
        is_sp_byze -> from_byze [label="No"]
        from_byze -> invoke

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, invoke
        }

        subgraph r2 {
            rank="same"
            is_app_byze, to_byze, is_sp_byze, from_byze
        }
    }



.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Response Body in Byze API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="Byze \nReceives \nResponse\nfrom\nService\nProvider"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_byze [label="App's Flavor\n==\nByze ?", shape=diamond]
        is_sp_byze [label="Flavor of \nService\nProvider\n==\nByze ?", shape=diamond]
        to_byze [label="convert\nResponse\nto\nByze\nFlavor"]
        from_byze [label="convert\nto\nApp's\nFlavor"]
        send [label="Send\nResponse\nin App's\nFlavor\nto App"]

        receive->is_same_flavor
        is_same_flavor->send [label="Yes"]
        is_same_flavor->is_sp_byze [label="No"]
        is_sp_byze -> is_app_byze [label="Yes"]
        is_sp_byze -> to_byze [label="No"]
        to_byze -> is_app_byze
        is_app_byze -> send [label="Yes"]
        is_app_byze -> from_byze [label="No"]
        from_byze -> send

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, send
        }

        subgraph r2 {
            rank="same"
            is_app_byze, to_byze, is_sp_byze, from_byze
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

混合调度策略以提高可用性
========================================================

``Byze`` 提供混合调度，即需要时，它将应用请求（经过必要转换）调度到远程替代方案 ``Byze Service Provider`` （通常是云服务）而不是本地。
当本地 AIPC 忙碌、当前 PC 不提供所需服务或用户想在云端使用 VIP 服务等情况下，这非常有帮助。

``Byze`` 通过遵循指定的 ``hybrid policy`` 来做出这样的调度决策。安装了 ``Byze`` 的 AIPC 具有系统级配置（参见 :doc:`/byze平台配置`），
它指定了所有可用的 ``Byze Service`` 以及它们对应的本地和远程 ``Byze Service Providers`` ，以及用于在这些提供者之间切换的默认 ``hybrid policy`` 。

此外，应用程序还可以覆盖平台配置中定义的默认 ``hybrid policy`` 。例如，应用程序可能强制使用云服务处理特定请求，此时可以在请求的 JSON 体中添加 ``hybrid_policy: always_remote`` 以发送。



.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Hybrid Scheduling"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 

     app[label="Application", fillcolor="#eeeeff"]
     byze[label="Byze to Dispatch - based on Hybrid Policy", fillcolor="#ffffcc"]
     local[label="Local Byze Service Provider", fillcolor="#ffcccc"]
     cloud[label="Remote Byze Service Provider", fillcolor="#ffcccc"]

     app -> byze

     byze -> local[style="dashed"]
     byze -> cloud[style="dashed"]

   }




.. _match_models:

模型匹配
========================================================

在许多情况下，应用程序可能希望指定要使用的首选模型，但底层 ``Byze Service Provider`` 要么不提供模型，要么提供的模型名称略有不同。

目前 ``Byze`` 提供了一种简单的机制，该机制试图从服务提供商中选择与所需模型最匹配的模型。这在未来可能会改变或发展。

首先，当定义可用的 ``Byze Service Provider`` 时，
:doc:`/byze平台配置` 也可以列出每个服务提供商的可用模型，作为其 :ref:`属性说明 <byze_service_provider_properties>` 的一部分。

然后，应用可以在请求中指定模型名称，例如，在请求的 JSON 体中使用 ``model: xx-7B`` 。 ``Byze`` 将在预期的模型和供应商提供的可用模型之间进行模糊匹配，并请求使用最相似的一个。
