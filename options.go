// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"time"

	"github.com/choria-io/go-choria/plugin"
)

const (
	// keys used in config file set by helper
	configKeySourceHost    = "machine_room.source.host"
	configKeySourceNatsJwt = "machine_room.source.nats_jwt"
	configKeyRole          = "machine_room.role"
	configKeySite          = "machine_room.site"

	// filesystem paths
	defaultServerStatusFile          = "/var/lib/choria/machine-room/status.json"
	defaultStorageDirectory          = "/var/lib/choria/machine-room"
	defaultReplicationStateDirectory = "/var/lib/choria/machine-room/replicator"

	// names of files stored in config dir
	defaultServerSeedFileName    = "server.seed"
	defaultServerJwtFileName     = "server.jwt"
	defaultMachineStore          = "machines"
	defaultProvisioningTokenFile = "provisioning.jwt"
	defaultFactsFile             = "instance.json"
	defaultNatsNkeyFile          = "nats.nkey"
	defaultNatsCredentialFile    = "nats.creds"
	defaultCaFile                = "ca.pem"
	defaultCertFile              = "cert.pem"
	defaultKeyFile               = "key.pem"

	// submission options
	defaultSubmissionSpool     = "/var/lib/choria/machine-room/submission"
	defaultSubmissionSpoolSize = 5000

	// default times and ports
	defaultFactsRefresh      = 10 * time.Minute
	defaultShutdownGrace     = 5 * time.Second
	defaultNetworkClientPort = 9222
)

type roOptions struct {
	opts Options
}

func (o roOptions) Name() string                        { return o.opts.Name }
func (o roOptions) Version() string                     { return o.opts.Version }
func (o roOptions) CommandPath() string                 { return o.opts.CommandPath }
func (o roOptions) MachineSigningKey() string           { return o.opts.MachineSigningKey }
func (o roOptions) FactsRefreshInterval() time.Duration { return o.opts.FactsRefreshInterval }
func (o roOptions) NoStandardFacts() bool               { return o.opts.NoStandardFacts }
func (o roOptions) NoMemoryFacts() bool                 { return o.opts.NoMemoryFacts }
func (o roOptions) NoSwapFacts() bool                   { return o.opts.NoSwapFacts }
func (o roOptions) NoCPUFacts() bool                    { return o.opts.NoCPUFacts }
func (o roOptions) NoDiskFacts() bool                   { return o.opts.NoDiskFacts }
func (o roOptions) NoHostFacts() bool                   { return o.opts.NoHostFacts }
func (o roOptions) NoNetworkFacts() bool                { return o.opts.NoNetworkFacts }
func (o roOptions) ConfigurationDirectory() string      { return o.opts.ConfigurationDirectory }
func (o roOptions) MachinesDirectory() string           { return o.opts.MachinesDirectory }
func (o roOptions) ProvisioningJWTFile() string         { return o.opts.ProvisioningJWTFile }
func (o roOptions) FactsFile() string                   { return o.opts.FactsFile }
func (o roOptions) SeedFile() string                    { return o.opts.ServerSeedFile }
func (o roOptions) JWTFile() string                     { return o.opts.ServerJWTFile }
func (o roOptions) StatusFile() string                  { return o.opts.ServerStatusFile }
func (o roOptions) SubmissionDirectory() string         { return o.opts.ServerSubmissionDirectory }
func (o roOptions) SubmissionSpoolSize() int            { return o.opts.ServerSubmissionSpoolSize }
func (o roOptions) StorageDirectory() string            { return o.opts.ServerStorageDirectory }
func (o roOptions) NatsNeySeedFile() string             { return o.opts.NatsNeySeedFile }
func (o roOptions) NatsCredentialsFile() string         { return o.opts.NatsCredentialsFile }
func (o roOptions) StartTime() time.Time                { return o.opts.StartTime }
func (o roOptions) ConfigBucketPrefix() string          { return o.opts.ConfigBucketPrefix }
func (o roOptions) Args() []string                      { return o.opts.Args }

func (o *Options) roCopy() *roOptions {
	return &roOptions{*o}
}

// Options holds configuration and runtime derived paths, members marked RO are set during CommonConfigure(), setting them has no effect
type Options struct {
	// Name is the name reported in --help and other output from the command line
	Name string `json:"name"`
	// Contact will be shown during --help
	Contact string `json:"contact"`
	// Help will be shown during --help as the main command help
	Help string `json:"help"`
	// Version will be reported in --version and elsewhere
	Version string `json:"version"`
	// MachineSigningKey hex encoded ed25519 key used to sign autonomous agents
	MachineSigningKey string `json:"machine_signing_key"`

	// optional below

	// FactsRefreshInterval sets an interval to refresh facts on, 10 minutes by default and cannot be less than 1 minute
	FactsRefreshInterval time.Duration `json:"facts_refresh_interval"`
	// ConfigBucketPrefix will replicate only a subset of keys from the backend to the site
	ConfigBucketPrefix string `json:"config_bucket_prefix"`
	// Plugins are additional plugins like autonomous agents to add to the build
	Plugins map[string]plugin.Pluggable `json:"-"`
	// AdditionalFacts will be called during fact generation and the result will be shallow merged with the standard facts
	AdditionalFacts FactsGenerator `json:"-"`
	// ReadyFunc is an optional function that will be called once provisioning completes and system is fully initialized
	ReadyFunc ReadyFunc `json:"-"`
	// Args are parsed instead of os.Args if Args is not nil
	Args []string `json:"-"`

	// facts related opt-outs
	// NoStandardFacts disables gathering all standard facts
	NoStandardFacts bool `json:"no_standard_facts,omitempty"`
	// NoMemoryFacts disables built-in memory fact gathering
	NoMemoryFacts bool `json:"no_memory_facts,omitempty"`
	// NoSwapFacts disables built-in swap facts gathering
	NoSwapFacts bool `json:"no_swap_facts,omitempty"`
	// NoCPUFacts disables built-in cpu facts gathering
	NoCPUFacts bool `json:"no_cpu_facts,omitempty"`
	// NoDiskFacts disables built-in disk facts gathering
	NoDiskFacts bool `json:"no_disk_facts,omitempty"`
	// NoHostFacts disables built-in host facts gathering
	NoHostFacts bool `json:"no_host_facts,omitempty"`
	// NoNetworkFacts disables built-in network interface facts gathering
	NoNetworkFacts bool `json:"no_network_facts,omitempty"`

	// Read only below...

	// ConfigurationDirectory is the directory the configuration file is stored in (RO)
	ConfigurationDirectory string `json:"configuration_directory"`
	// MachinesDirectory is where autonomous agents are stored (RO)
	MachinesDirectory string `json:"machines_directory"`
	// ProvisioningJWTFile is the path to provisioning jwt file, defaults to provisioning.jwt in the options dir (RO)
	ProvisioningJWTFile string `json:"provisioning_jwt_file"`
	// FactsFile is the path to the facts file which default to instance.json in the options dir (RO)
	FactsFile string `json:"facts_file"`
	// ServerSeedFile is the path to the server seed file that will exist after provisioning (RO)
	ServerSeedFile string `json:"server_seed_file"`
	// ServerJWTFile is the path to the server jwt file that will exist after provisioning (RO)
	ServerJWTFile string `json:"server_jwt_file"`
	// ServerStatusFile is where the server will regularly write its status (RO)
	ServerStatusFile string `json:"server_status_file"`
	// ServerSubmissionDirectory is the directory holding the submission spool (RO)
	ServerSubmissionDirectory string `json:"server_submission_directory"`
	// ServerSubmissionSpoolSize is the maximum size of the submission spool (RO)
	ServerSubmissionSpoolSize int `json:"server_submission_spool_size"`
	// CommandPath is the path to the command being run, defaults to argv[0] (RO)
	CommandPath string `json:"command_path"`
	// ServerStorageDirectory the directory where state is stored (RO)
	ServerStorageDirectory string `json:"server_storage_directory"`
	// NatsNeySeedFile is a path to a nkey seed created at start
	NatsNeySeedFile string `json:"nats_ney_seed_file"`
	// NatsCredentialsFile is a path to the nats credentials file holding data received during provisioning
	NatsCredentialsFile string `json:"nats_credentials_file"`
	// StartTime the time the process started (RO)
	StartTime time.Time `json:"start_time"`
}
