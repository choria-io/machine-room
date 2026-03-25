## Example Machine Room-based SaaS

This is an evolving example setup of a machine room-based SaaS. It is based on a previous example that has been lost.

This re-creates that example and will slowly evolve it to be a bit more realistic.

## Overview

This diagram shows the overall architecture of the example. 

1. The goal is to create a SaaS that can manage software deployments in many customer sites (2) here we have Dashboards, API etc
2. The customer sites are connected to a central server (1). One server is designated `leader` and runs a Choria Broker that hosts some streams and replicate data to/from the backend.
3. The customer sites are provisioned using a Choria Provisioner that interacts with the backend via API calls to get credentials and configuration
4. Software managed in the customer site is via Autonomous Agents which are downloaded by the customer site from a artifact server via HTTPS

![Overview](overview.png)


At present the Dashboard, Database and API is not in this demo - that is inherently specific to the SaaS being built. Data ends in the `SaaS NATS` ready for consumption.

Machine Room presents the `SaaS NATS` as the only interaction point with the customer sites, the management portal consumes streams for node state and events and write Key-Value data to capture configuration values and desired plugins to deploy to a site.

At present no RPC is supported to the Customer Sites.

## Using

Run `docker compose up --build` which will build the agent container (example/agent) and start the entire setup.

When done do `docker compose down -v` to shut everything down.

At start the `example/setup.sh` is run to create all the credentials and then while running the `shell` instance can be accessed and it has all the generated files, configurations etc in `/machine-room`

## Creating and deploying plugins

TODO