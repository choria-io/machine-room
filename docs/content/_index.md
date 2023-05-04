+++
weight = 5
+++

## Overview

This is a tool that allows automation backplanes to be built that are specifically tailored to the needs of managed SaaS
software vendors.

![Overview](/machine-room-overview.png)

Vendors who wish to build a SaaS offerings that allow their software to be installed and managed on private, customer owned, infrastructure
can use this to build the client-side software agent quickly and easily.

This tool takes care of:

 * Early provisioning of a customer into your SaaS
 * Management of applications using [Choria Autonomous Agents](https://choria.io/docs/autoagents/)
 * Data streaming, events and metrics communicated back to the SaaS using [cloudevents](https://cloudevents.io/) format
 * Provides various orchestration primitives for managing multi node cluster rollouts
 * Over the air deployment of management control loops at runtime and at tremendous scale and performance

Using this a SaaS can go from zero to code running safely in a customer environment with less than 100 lines of code, 
allowing you to focus on your core business needs.

A single customer "site" can scale to 10s of thousands of nodes, and we support complex topologies and orchestration.

## Example

An example speaks volumes, the code below is a complete setup of a Machine Room agent:

```golang
func main() {
	app, err := machineroom.New(machineroom.Options{
		Name:    "example-manager",
		Version: "0.0.1",
		Contact: "info@example.net",
		Help:    "Example Manager", 
		
		// SaaS identity that verifies all plugins, ota updates and more
		MachineSigningKey: "b217b9c7574.....3522da5c3b82",
	})
	panicIfError(err)

	panicIfError(app.Run(context.Background()))
}
```

This code, when deployed to a customer, will:

 * Read a SaaS-issued JWT holding provisioning information
 * Connect to the provisioning network and enroll with the SaaS
 * Restart into provisioned state:
   * Start a local Choria Broker
   * Create a number of local spool streams
   * Set up Stream Replicator to copy data between customer and SaaS
   * Regularly publish server metadata, configuration and life-cycle events
   * Download Autonomous Agents from the SaaS and health check them continuously

From this point any autonomous agent actions can be used to manage software on the nodes.  The SaaS can upgrade and
downgrade customer sites at will.

If the customer site is offline desired state remains in the customer side and events are spooled.  Once the customer
is back in on-line the events will spool back to the SaaS.

Here's output from the above `main.go`:

```nohighlight
[root@managed /]# example-manager
usage: example-manager [<flags>] <command> [<args> ...]

Acme Manager

Commands:
  run    Runs the management agent
  reset  Restores the agent to factory defaults

Pass --help to see global flags applicable to this command.
```

And this is the agent running on a node:

```nohighlight
[root@managed /]# ps -auxw
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root           1  0.2  0.1 2312836 45360 ?       Ssl  12:35   0:21 /usr/bin/example-manager
```

## Status

This is a work in progress, while we are combining existing capabilities (Broker, Server, Stream Replicator and more) into
to this framework to build this capability the exact design and affordances that a tool like this should expose to the 
SaaS developer.
