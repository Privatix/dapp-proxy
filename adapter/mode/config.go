package mode

import "github.com/privatix/dappctrl/util/log"

// Keys in product configuration.
const (
	productAlterID = "alterId"
	productAddress = "address"
	productPort    = "port"
)

// AgentConfig is agent adapter configuration.
type AgentConfig struct {
	FileLog *log.FileConfig
	V2Ray   V2RayAgentConfig
	Sess    SessConfig
	Monitor MonitorConfig
}

// ValidAgentConf returns true if config has proper v2ray config.
func ValidAgentConf(c *AgentConfig) bool {
	// alter id cannot be 0
	return c.V2Ray.AlterID > 0
}

// V2RayAgentConfig is agent v2ray config.
type V2RayAgentConfig struct {
	AlterID     uint32
	API         string
	InboundTag  string
	InboundPort uint
}

// ClientConfig is client adapter configuration.
type ClientConfig struct {
	FileLog *log.FileConfig
	V2Ray   V2RayClientConfig
	Sess    SessConfig
	Monitor MonitorConfig
}

// V2RayClientConfig is client v2ray config.
type V2RayClientConfig struct {
	API        string
	InboundTag string
}

// SessConfig is configariotion to connect to session server.
type SessConfig struct {
	Endpoint string
	Origin   string
	Product  string
	Password string
}

// MonitorConfig monitor configuration.
type MonitorConfig struct {
	CountPeriod uint // in seconds.
}
