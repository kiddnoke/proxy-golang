package shadowsocks

type SSconfig struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	Timeout    int    `json:"timeout"`
	Limit      int64  `json:"limit"`
	Method     string `json:"method"`
	Password   string `json:"password"`
	Expiration int64  `json:"expiration"`
	ServerPort int    `json:"server_port"`
}
