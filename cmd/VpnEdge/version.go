package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

var (
	BuildDate    string
	BuildVersion string
	BuildBranch  string
)

func init() {
	if BuildVersion == "" || BuildDate == "" {
		BuildVersion = "Debug"
		BuildDate = "Debug"
		BuildBranch = "Debug"
	}
	log.Printf("BuildVersion [%s]", BuildVersion)
	log.Printf("BuildDate [%s]", BuildDate)
	log.Printf("BuildBranch [%s]", BuildBranch)
}
func Generate() {
	log.Printf("生成pm2版本文件")
	var writeString = fmt.Sprintf("{\"version\":\"%s\"}", BuildBranch+"-"+BuildDate)
	filename := "./package.json"
	var d1 = []byte(writeString)
	ioutil.WriteFile(filename, d1, 0666)
	return
}
