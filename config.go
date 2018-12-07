package shadowsocks

type Config struct {
	Uid        int64  `json:"uid"`
	Sid        int64  `json:"sid"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Timeout    int64  `json:"timeout"`
	Expiration int64  `json:"expiration"`
	Limit      int64  `json:"limit"`
}
