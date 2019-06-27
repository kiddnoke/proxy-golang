package softether

import (
	"github.com/kiddnoke/SoftetherGo"
	"log"
)

var API *softetherApi.API

var SoftHost string
var SoftPort int
var SoftPassword string

var ServerCert string
var RemoteAccess string
var DDnsHostName string
var Ipv4Address string

var OpenVpnServicePort int

func Init() {
	if checkSoftetherIsFirst() == true {
		if err := softetherFirstInit(SoftPassword); err != nil {
			panic(err)
		}
	} else {
		API = softetherApi.NewAPI(SoftHost, SoftPort, SoftPassword)
		if err := API.HandShake(); err != nil {
			panic(err)
		}
	}

	//
	if remoteaccessfile, err := API.GetOpenVpnRemoteAccess(); err != nil {
		panic(err)
	} else {
		RemoteAccess = remoteaccessfile
	}
	//
	if cert, err := API.GetServerCert(); err != nil {
		panic(err)
	} else {
		ServerCert = cert
	}
	//
	if hostname, ipv4, err := API.GetDDnsHostName(); err != nil {
		panic(err)
	} else {
		DDnsHostName = hostname
		Ipv4Address = ipv4
	}
	Cron()
	log.Println("Softether Init Success")
}
func createDefaultHub() {

}
func createDefalutGroup() {

}
func checkSoftetherIsFirst() bool {
	API = softetherApi.NewAPI(SoftHost, SoftPort, "")
	if err := API.HandShake(); err != nil {
		API = nil
		return false
	} else {
		return true
	}
}
func softetherFirstInit(password string) error {
	if _, err := API.SetServerPassword(password); err != nil {
		return err
	} else {
		return nil
	}
}
