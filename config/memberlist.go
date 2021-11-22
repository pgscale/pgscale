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

import (
	"os"
	"time"

	"github.com/hashicorp/memberlist"
)

// prepareMemberlistConfig creates a new *memberlist.Config by parsing olricd.yaml
func prepareMemberlistConfig(c Olric, mc *memberlist.Config) (*memberlist.Config, error) {
	var err error
	if c.Memberlist.BindAddr == "" {
		name, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		c.Memberlist.BindAddr = name
	}
	mc.BindAddr = c.Memberlist.BindAddr
	mc.BindPort = c.Memberlist.BindPort

	if c.Memberlist.EnableCompression != nil {
		mc.EnableCompression = *c.Memberlist.EnableCompression
	}

	if c.Memberlist.TCPTimeout != nil {
		mc.TCPTimeout, err = time.ParseDuration(*c.Memberlist.TCPTimeout)
		if err != nil {
			return nil, err
		}
	}

	if c.Memberlist.IndirectChecks != nil {
		mc.IndirectChecks = *c.Memberlist.IndirectChecks
	}

	if c.Memberlist.RetransmitMult != nil {
		mc.RetransmitMult = *c.Memberlist.RetransmitMult
	}

	if c.Memberlist.SuspicionMult != nil {
		mc.SuspicionMult = *c.Memberlist.SuspicionMult
	}

	if c.Memberlist.PushPullInterval != nil {
		mc.PushPullInterval, err = time.ParseDuration(*c.Memberlist.PushPullInterval)
		if err != nil {
			return nil, err
		}
	}

	if c.Memberlist.ProbeTimeout != nil {
		mc.ProbeTimeout, err = time.ParseDuration(*c.Memberlist.ProbeTimeout)
		if err != nil {
			return nil, err
		}
	}
	if c.Memberlist.ProbeInterval != nil {
		mc.ProbeInterval, err = time.ParseDuration(*c.Memberlist.ProbeInterval)
		if err != nil {
			return nil, err
		}
	}

	if c.Memberlist.GossipInterval != nil {
		mc.GossipInterval, err = time.ParseDuration(*c.Memberlist.GossipInterval)
		if err != nil {
			return nil, err
		}
	}
	if c.Memberlist.GossipToTheDeadTime != nil {
		mc.GossipToTheDeadTime, err = time.ParseDuration(*c.Memberlist.GossipToTheDeadTime)
		if err != nil {
			return nil, err
		}
	}

	if c.Memberlist.AdvertiseAddr != nil {
		mc.AdvertiseAddr = *c.Memberlist.AdvertiseAddr
	}

	if c.Memberlist.AdvertisePort != nil {
		mc.AdvertisePort = *c.Memberlist.AdvertisePort
	} else {
		mc.AdvertisePort = mc.BindPort
	}

	if c.Memberlist.SuspicionMaxTimeoutMult != nil {
		mc.SuspicionMaxTimeoutMult = *c.Memberlist.SuspicionMaxTimeoutMult
	}

	if c.Memberlist.DisableTCPPings != nil {
		mc.DisableTcpPings = *c.Memberlist.DisableTCPPings
	}

	if c.Memberlist.AwarenessMaxMultiplier != nil {
		mc.AwarenessMaxMultiplier = *c.Memberlist.AwarenessMaxMultiplier
	}

	if c.Memberlist.GossipNodes != nil {
		mc.GossipNodes = *c.Memberlist.GossipNodes
	}
	if c.Memberlist.GossipVerifyIncoming != nil {
		mc.GossipVerifyIncoming = *c.Memberlist.GossipVerifyIncoming
	}
	if c.Memberlist.GossipVerifyOutgoing != nil {
		mc.GossipVerifyOutgoing = *c.Memberlist.GossipVerifyOutgoing
	}

	if c.Memberlist.DNSConfigPath != nil {
		mc.DNSConfigPath = *c.Memberlist.DNSConfigPath
	}

	if c.Memberlist.HandoffQueueDepth != nil {
		mc.HandoffQueueDepth = *c.Memberlist.HandoffQueueDepth
	}
	if c.Memberlist.UDPBufferSize != nil {
		mc.UDPBufferSize = *c.Memberlist.UDPBufferSize
	}
	return mc, nil
}
