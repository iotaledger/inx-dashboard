---
description: INX-Dashboard is a node configuration dashboard for node owners.
image: /img/Banner/banner_hornet.png
keywords:
- IOTA Node
- Hornet Node
- INX
- Dashboard
- IOTA
- Shimmer
- Node Software
- Welcome
- explanation
---

# Welcome to INX-Dashboard

INX-Dashboard provides a web GUI tool to help you manage your node. It provides an overview of the node's status and health and allows you to choose its peers, as well as the enabled plugins. It has a built-in explorer of all nodes on the network and visualizer for activity on the Tangle.

## Setup

We recommend you to use the [Docker images](https://hub.docker.com/r/iotaledger/inx-dashboard).
These images are also used in the [Docker setup](http://wiki.iota.org/hornet/develop/how_tos/using_docker) of Hornet.

## Configuration

The dashboard connects to the local Hornet instance by default.
It exposes the web GUI on port `8081` by default.

The dashboard provides built-in access control that you can [configure](./configuration.md#dashboard_auth) by setting the values for `dashboard.auth.passwordHash` and `dashboard.auth.passwordSalt`.

:::note

You can use `hornet tool pwd-hash` to generate credentials that you could use in the configuration file.

:::

Check the [Set Dashboard Credentials](http://wiki.iota.org/hornet/develop/how_tos/using_docker#4-set-dashboard-credentials) section of the recommended setup for more details.

You can find all the configuration options in the [Configuration section](configuration.md).

## Source Code

The source code of the project is available on [GitHub](https://github.com/iotaledger/inx-dashboard).