package main

import "log"

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
