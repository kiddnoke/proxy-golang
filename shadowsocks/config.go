package shadowsocks

type SSconfig struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	Timeout    int    `json:"timeout"`
	Method     string `json:"method"`
	Password   string `json:"password"`
	Expire     int64  `json:"expire"`
	ServerPort int    `json:"server_port"`
	Limitup    int    `json:"limitup"`
	Limitdown  int    `json:"limitdown"`
}
