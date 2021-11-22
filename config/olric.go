// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

type OlricLogging struct {
	Verbosity int32  `hcl:"verbosity"`
	Level     string `hcl:"level"`
	Output    string `hcl:"output"`
}

type Client struct {
	DialTimeout  string  `hcl:"dial_timeout"`
	ReadTimeout  string  `hcl:"read_timeout"`
	WriteTimeout string  `hcl:"write_timeout"`
	KeepAlive    string  `hcl:"keep_alive"`
	MinConn      int     `hcl:"min_conn"`
	MaxConn      int     `hcl:"max_conn"`
	PoolTimeout  *string `hcl:"pool_timeout"`
}

type Memberlist struct {
	Environment             string   `hcl:"environment"` // required
	BindAddr                string   `hcl:"bind_addr"`   // required
	BindPort                int      `hcl:"bind_port"`   // required
	Interface               *string  `hcl:"interface"`
	EnableCompression       *bool    `hcl:"enable_compression"`
	JoinRetryInterval       string   `hcl:"join_retry_interval"` // required
	MaxJoinAttempts         int      `hcl:"max_join_attempts"`   // required
	Peers                   []string `hcl:"peers"`
	IndirectChecks          *int     `hcl:"indirect_checks"`
	RetransmitMult          *int     `hcl:"retransmit_mult"`
	SuspicionMult           *int     `hcl:"suspicion_mult"`
	TCPTimeout              *string  `hcl:"tcp_timeout"`
	PushPullInterval        *string  `hcl:"push_pull_interval"`
	ProbeTimeout            *string  `hcl:"probe_timeout"`
	ProbeInterval           *string  `hcl:"probe_interval"`
	GossipInterval          *string  `hcl:"gossip_interval"`
	GossipToTheDeadTime     *string  `hcl:"gossip_to_the_dead_time"`
	AdvertiseAddr           *string  `hcl:"advertise_addr"`
	AdvertisePort           *int     `hcl:"advertise_port"`
	SuspicionMaxTimeoutMult *int     `hcl:"suspicion_max_timeout_mult"`
	DisableTCPPings         *bool    `hcl:"disable_tcp_pings"`
	AwarenessMaxMultiplier  *int     `hcl:"awareness_max_multiplier"`
	GossipNodes             *int     `hcl:"gossip_nodes"`
	GossipVerifyIncoming    *bool    `hcl:"gossip_verify_incoming"`
	GossipVerifyOutgoing    *bool    `hcl:"gossip_verify_outgoing"`
	DNSConfigPath           *string  `hcl:"dns_config_path"`
	HandoffQueueDepth       *int     `hcl:"handoff_queue_depth"`
	UDPBufferSize           *int     `hcl:"udp_buffer_size"`
}

type StorageEngines struct {
	Plugins *[]string                    `hcl:"plugins"`
	Engines map[string]map[string]string `hcl:"engines"`
}

type Olric struct {
	Name                     *string        `hcl:"name"`
	BindAddr                 string         `hcl:"bind_addr"`
	BindPort                 int            `hcl:"bind_port"`
	Interface                *string        `hcl:"interface"`
	ReplicationMode          int            `hcl:"replication_mode"`
	PartitionCount           uint64         `hcl:"partition_count"`
	LoadFactor               *float64       `hcl:"load_factor"`
	Serializer               string         `hcl:"serializer"`
	KeepAlivePeriod          string         `hcl:"keepalive_period"`
	BootstrapTimeout         string         `hcl:"bootstrap_timeout"`
	ReplicaCount             int            `hcl:"replica_count"`
	WriteQuorum              int            `hcl:"write_quorum"`
	ReadQuorum               int            `hcl:"read_quorum"`
	ReadRepair               bool           `hcl:"read_repair"`
	MemberCountQuorum        int32          `hcl:"member_count_quorum"`
	RoutingTablePushInterval string         `hcl:"routing_table_push_interval"`
	Client                   Client         `hcl:"client,block"`
	Logging                  OlricLogging   `hcl:"logging,block"`
	Memberlist               Memberlist     `hcl:"memberlist,block"`
	StorageEngines           StorageEngines `hcl:"storage_engines,block"`
}
