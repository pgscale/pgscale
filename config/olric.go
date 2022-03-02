package config

type OlricLogging struct {
	Verbosity int32  `yaml:"verbosity"`
	Level     string `yaml:"level"`
	Output    string `yaml:"output"`
}

type Client struct {
	DialTimeout  string `yaml:"dialTimeout"`
	ReadTimeout  string `yaml:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout"`
	KeepAlive    string `yaml:"keepAlive"`
	MinConn      int    `yaml:"minConn"`
	MaxConn      int    `yaml:"maxConn"`
	PoolTimeout  string `yaml:"poolTimeout"`
}

type Memberlist struct {
	Environment             string   `yaml:"environment"` // required
	BindAddr                string   `yaml:"bindAddr"`    // required
	BindPort                int      `yaml:"bindPort"`    // required
	Interface               *string  `yaml:"interface"`
	EnableCompression       *bool    `yaml:"enableCompression"`
	JoinRetryInterval       string   `yaml:"joinRetryInterval"` // required
	MaxJoinAttempts         int      `yaml:"maxJoinAttempts"`   // required
	Peers                   []string `yaml:"peers"`
	IndirectChecks          *int     `yaml:"indirectChecks"`
	RetransmitMult          *int     `yaml:"retransmitMult"`
	SuspicionMult           *int     `yaml:"suspicionMult"`
	TCPTimeout              *string  `yaml:"tcpTimeout"`
	PushPullInterval        *string  `yaml:"pushPullInterval"`
	ProbeTimeout            *string  `yaml:"probeTimeout"`
	ProbeInterval           *string  `yaml:"probeTnterval"`
	GossipInterval          *string  `yaml:"gossipInterval"`
	GossipToTheDeadTime     *string  `yaml:"gossipToTheDeadTime"`
	AdvertiseAddr           *string  `yaml:"advertiseAddr"`
	AdvertisePort           *int     `yaml:"advertisePort"`
	SuspicionMaxTimeoutMult *int     `yaml:"suspicionMaxTimeoutMult"`
	DisableTCPPings         *bool    `yaml:"disableTCPPings"`
	AwarenessMaxMultiplier  *int     `yaml:"awarenessMaxMultiplier"`
	GossipNodes             *int     `yaml:"gossipNodes"`
	GossipVerifyIncoming    *bool    `yaml:"gossipVerifyIncoming"`
	GossipVerifyOutgoing    *bool    `yaml:"gossipVerifyOutgoing"`
	DNSConfigPath           *string  `yaml:"dnsConfigPath"`
	HandoffQueueDepth       *int     `yaml:"handoffQueueDepth"`
	UDPBufferSize           *int     `yaml:"udpBufferSize"`
}

type StorageEngines struct {
	Plugins *[]string                    `yaml:"plugins"`
	Engines map[string]map[string]string `yaml:"engines"`
}

type Olric struct {
	Name                     *string        `yaml:"name"`
	BindAddr                 string         `yaml:"bindAddr"`
	BindPort                 int            `yaml:"bindPort"`
	Interface                *string        `yaml:"interface"`
	ReplicationMode          int            `yaml:"replicationMode"`
	PartitionCount           uint64         `yaml:"partitionCount"`
	LoadFactor               *float64       `yaml:"loadFactor"`
	Serializer               string         `yaml:"serializer"`
	KeepAlivePeriod          string         `yaml:"keepalivePeriod"`
	BootstrapTimeout         string         `yaml:"bootstrapTimeout"`
	ReplicaCount             int            `yaml:"replicaCount"`
	WriteQuorum              int            `yaml:"writeQuorum"`
	ReadQuorum               int            `yaml:"readQuorum"`
	ReadRepair               bool           `yaml:"readRepair"`
	MemberCountQuorum        int32          `yaml:"memberCountQuorum"`
	RoutingTablePushInterval string         `yaml:"routingTablePushInterval"`
	Client                   Client         `yaml:"client"`
	Logging                  OlricLogging   `yaml:"logging"`
	Memberlist               Memberlist     `yaml:"memberlist"`
	StorageEngines           StorageEngines `yaml:"storageEngines"`
}
