package main

import (
	ss "../../shadowsocks"
	"flag"
)

func main() {
	var manager_port int
	var core int
	var loggerlevel int
	var logFile string
	var centerhost string
	var centerport int
	flag.IntVar(&loggerlevel, "loggerlevel", 1, "logger level ")
	flag.StringVar(&logFile, "log-file", "./log", "log-file path")
	flag.IntVar(&manager_port, "manager-port", 8000, "manager deamon udp port")
	flag.IntVar(&core, "core", 0, "maximum number of CPU cores to use, default is determinied by Go runtime")
	flag.IntVar(&centerport, "P", 7001, "center port ")
	flag.StringVar(&centerhost, "H", "localhost", "center host ")
	flag.Parse()
	m := ss.NewManager(8000)
	ss.ManagerDaemon(m)
}
