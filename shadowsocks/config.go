package shadowsocks

type SSconfig struct {
	Uid           int    `json:"uid"`
	Sid           int    `json:"sid"`
	Timeout       int    `json:"timeout"`
	Method        string `json:"method"`
	Password      string `json:"password"`
	Expire        int64  `json:"expire"`
	ServerPort    int    `json:"server_port"`
	Currlimitup   int    `json:"currlimitup"`
	Currlimitdown int    `json:"currlimitdown"`
	Nextlimitup   int    `json:"nextlimitup"`
	Nextlimitdown int    `json:"nextlimitdown"`
	Remain        int    `json:"remain"`
}
