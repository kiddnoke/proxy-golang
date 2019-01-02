package manager

type Config struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	Timeout    int    `json:"timeout"`
	Method     string `json:"method"`
	Password   string `json:"password"`
	Expire     int64  `json:"expire"`
	ServerPort int    `json:"server_port"`
	Limit      int    `json:"limit"`
	Overflow   int64  `json:"overflow"`
}
