package shadowsocks

type Config struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Timeout    int    `json:"timeout"`
	Expiration int    `json:"expiration"`
	Limit    int64    `json:"limit"`
}
