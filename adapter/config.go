package adapter

type config struct {
	V2Ray   v2rayConfig
	Sess    sessConfig
	Monitor monitorConfig
}

type v2rayConfig struct {
	AlterID    uint32
	API        string
	InboundTag string
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
