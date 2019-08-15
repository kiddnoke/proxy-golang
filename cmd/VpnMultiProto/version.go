package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"proxy-golang/common"
	"strconv"
)

var (
	BuildDate    string
	BuildVersion string
	BuildBranch  string
	LoggerLevel  string
)

func init() {
	if BuildVersion == "" || BuildDate == "" {
		BuildVersion = "Debug"
		BuildDate = "Debug"
		BuildBranch = "Debug"
		LoggerLevel = strconv.Itoa(common.LOG_DEBUG)
	}
	log.Printf("BuildVersion [%s]", BuildVersion)
	log.Printf("BuildDate [%s]", BuildDate)
	log.Printf("BuildBranch [%s]", BuildBranch)
	i_loggerlevel, _ := strconv.Atoi(LoggerLevel)
	common.SetLoggerDefaultLevel(i_loggerlevel)
	commomLogLevel, _ := common.GetLoggerDefaultLevel()
	log.Printf("LoggerLevel %s", commomLogLevel)
}
func GeneratePm2ConfigFile() (err error) {
	log.Printf("生成pm2版本文件")
	var writeString = fmt.Sprintf("{\"version\":\"%s\"}", BuildBranch+"-"+BuildVersion+"-"+BuildDate)
	filename := "./package.json"
	var d1 = []byte(writeString)
	err = ioutil.WriteFile(filename, d1, 0666)
	return
}
