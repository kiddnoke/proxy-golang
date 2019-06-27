package multiprotocol

type Config struct {
	// v1.1
	Uid                   int64  `json:"uid"`
	Sid                   int64  `json:"sid"`
	ServerPort            int    `json:"server_port"`
	Method                string `json:"method"`
	Password              string `json:"password"`
	CurrLimitUp           int    `json:"currlimitup"`
	CurrLimitDown         int    `json:"currlimitdown"`
	Timeout               int64  `json:"timeout"`
	Expire                int64  `json:"expire"`
	BalanceNotifyDuration int    `json:"balancenotifytime"`
	// v1.1.1
	SnId             int64   `json:"sn_id"`
	AppVersion       string  `json:"app_version"`
	UserType         string  `json:"user_type"`
	CarrierOperators string  `json:"carrier_operators"`
	Os               string  `json:"os"`
	DeviceId         string  `json:"device_id"`
	UsedTotalTraffic int64   `json:"used_total_traffic" unit:"kb"`
	LimitArray       []int64 `json:"limit_array" unit:"kb"`
	FlowArray        []int64 `json:"flow_array" unit:"kb"`
	// NovaPro
	AppId       int64  `json:"app_id"`
	NetworkType string `json:"network_type"`
	//
	Ip    string `json:"ip"`
	State string `json:"state"`
	// OpenVpn
	Protocol     string `json:"protocol"`
	RemoteAccess string `json:"remote_access"`
	ServerCert   string `json:"server_cert"`
	Ipv4Address  string `json:"ipv4_address"`
}
