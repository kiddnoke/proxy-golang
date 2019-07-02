package ss

var methodtable = map[int]string{
	0: "aes-128-cfb", 1: "aes-192-cfb", 2: "aes-256-cfb"}

func GenCipherMethod(index int) string {
	if 0 <= index && index < 2 {
		return methodtable[index]
	}
	return "aes-128-cfb"
}
