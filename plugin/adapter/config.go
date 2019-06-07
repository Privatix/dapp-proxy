package adapter

import "github.com/privatix/dappctrl/util/log"

// Keys in product configuration.
const (
	productAlterID = "alterId"
	productAddress = "address"
	productPort    = "port"
)

// Config is adapter configuration.
type Config struct {
	FileLog *log.FileConfig
	V2Ray   V2RayConfig
	Sess    SessConfig
	Monitor MonitorConfig

	// Only for clients.

	// ConfigureProxyScript can configure operating system to use
	// or stop using sock5 proxy.
	ConfigureProxyScript string
	// ProxyBackupFile is a location of a file used to backup proxy configs before
	// making any change.
	ProxyBackupFile string
	// ChannelDir used to place active channel files.
	ChannelDir string
	// ProxyPort is the port proxy is listening on '127.0.0.1'
	ProxyPort int
}

// ValidAgentConf returns true if config has proper v2ray config.
func ValidAgentConf(c *Config) bool {
	return c.V2Ray.InboundPort > 0
}

// V2RayConfig is v2ray config.
type V2RayConfig struct {
	AlterID     uint32
	API         string
	InboundTag  string
	InboundPort uint
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
