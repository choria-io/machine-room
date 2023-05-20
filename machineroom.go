// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"time"

	"github.com/choria-io/fisk"
	"github.com/sirupsen/logrus"
)

// Instance is an instance of the Choria Machine Room Agent
type Instance interface {
	// Run starts running the command line
	Run(ctx context.Context) error
	// Application allows adding additional commands to the CLI application that will be built
	Application() *fisk.Application
	// CommonConfigure performs basic setup that a command added using Application() might need
	CommonConfigure() (RuntimeOptions, *logrus.Entry, error)
}

// RuntimeOptions provides read only access to run-time state and configuration
type RuntimeOptions interface {
	// Name is the configured application name
	Name() string
	// Version is the running version
	Version() string
	// CommandPath is the full path to the command being executed
	CommandPath() string
	// MachineSigningKey is the ed25519 public key used to sign autonomous agents and other items
	MachineSigningKey() string
	// FactsRefreshInterval is the frequency facts will be refreshed on disk
	FactsRefreshInterval() time.Duration
	// NoStandardFacts indicates if all built-in facts are disabled
	NoStandardFacts() bool
	// NoMemoryFacts indicates if built-in memory facts will be gathered
	NoMemoryFacts() bool
	// NoSwapFacts indicates if built-in swap facts will be gathered
	NoSwapFacts() bool
	// NoCPUFacts indicates if built-in cpu facts will be gathered
	NoCPUFacts() bool
	// NoDiskFacts indicates if built-in disk facts will be gathered
	NoDiskFacts() bool
	// NoHostFacts indicates if built-in host facts will be gathered
	NoHostFacts() bool
	// NoNetworkFacts indicates if built-in network facts will be gathered
	NoNetworkFacts() bool
	// ConfigurationDirectory is the path where configuration and other runtime files will be stored
	ConfigurationDirectory() string
	// MachinesDirectory is the directory where autonomous agents will be stored
	MachinesDirectory() string
	// ProvisioningJWTFile is the file issued by the SaaS provider used during provisioning
	ProvisioningJWTFile() string
	// FactsFile is a file holding instance data
	FactsFile() string
	// SeedFile is a ed25519 seed issued by the Choria Organization Issuer during provisioning
	SeedFile() string
	// JWTFile is the JWT file issued during provisioning
	JWTFile() string
	// StatusFile is a regularly updated file holding internal status of the Choria Backplane Server
	StatusFile() string
	// SubmissionDirectory is a spool directory that will hold messages submitted via Choria Submit
	SubmissionDirectory() string
	// SubmissionSpoolSize is the maximum size of the spool
	SubmissionSpoolSize() int
	// StorageDirectory is where JetStream and other state is kept
	StorageDirectory() string
	// NatsNeySeedFile is a NKey created during provisioning that could optionally be used to authenticate to the SaaS
	NatsNeySeedFile() string
	// NatsCredentialsFile is a NATS credential that, if provisioning signed a nats JWT, will hold a valid cred for accessing the SaaS backend
	NatsCredentialsFile() string
	// StartTime is the time this instance was started
	StartTime() time.Time
	// ConfigBucketPrefix will replicate only a subset of keys from the backend to the site
	ConfigBucketPrefix() string
}

// FactsGenerator gathers facts
type FactsGenerator func(ctx context.Context, opts RuntimeOptions, log *logrus.Entry) (map[string]any, error)

// ReadyFunc is a custom function that will be called after provisioning and initialization
type ReadyFunc func(ctx context.Context, opts RuntimeOptions, log *logrus.Entry)

// New creates a new machine room agent instance based on options
func New(o Options) (Instance, error) {
	return newCli(o)
}
