package adapter

// Keys in product configuration.
const (
	productAlterID = "alterId"
	productAddress = "address"
	productPort    = "port"
)

type agentConfig struct {
	V2Ray   v2rayAgentConfig
	Sess    sessConfig
	Monitor monitorConfig
}

type v2rayAgentConfig struct {
	AlterID     uint32
	API         string
	InboundTag  string
	InboundPort uint
}

type clientConfig struct {
	V2Ray   v2rayAgentConfig
	Sess    sessConfig
	Monitor monitorConfig
}

type v2rayClientConfig struct {
	API        string
	InboundTag string
	ExecPath   string
}

type sessConfig struct {
	Endpoint string
	Origin   string
	Product  string
	Password string
}

type monitorConfig struct {
	CountPeriod uint // in seconds.
}
