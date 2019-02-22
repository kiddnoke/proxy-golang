package main

import "log"

var (
	BuildDate    string
	BuildVersion string
)

func init() {
	if BuildVersion == "" || BuildDate == "" {
		BuildVersion = "Debug"
		BuildDate = "Debug"
	}
	log.Printf("version [%s]", BuildVersion)
	log.Printf("BuildDate [%s]", BuildDate)
}
