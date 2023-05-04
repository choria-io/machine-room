+++
weight = 5
+++

# Overview

This is a tool that allows automation backplanes to be built that are specifically tailored to the needs of managed SaaS
software vendors.

If you are a vendor who wish to build a SaaS that allows your software to be installed and managed on private, customer 
owned, infrastructure.

![Overview](/machine-room-overview.png)

This tool takes care of:

 * Early provisioning of a customer into your SaaS
 * Management of applications using [Choria Autonomous Agents](https://choria.io/docs/autoagents/)
 * Data streaming, events and metrics communicated back to the SaaS using [cloudevents](https://cloudevents.io/) format
 * Provides various orchestration primitives for managing multi node cluster rollouts
 * Over the air deployment of management control loops at runtime and at tremendous scale and performance

Using this a SaaS can go from zero to code running safely in a customer environment with less than 100 lines of code, 
allowing you to focus on your core business needs.

A single customer "site" can scale to 10s of thousands of nodes, and we support complex topologies and orchestration.

## Status

This is a work in progress, while we are combining existing capabilities (Broker, Server, Stream Replicator and more) into
to this framework to build this capability the exact design and affordances that a tool like this should expose to the 
SaaS developer.
