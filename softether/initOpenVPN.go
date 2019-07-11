package softether

import (
	"bufio"
	"fmt"
	"github.com/kiddnoke/SoftetherGo"
	"log"
	"strconv"
	"strings"
)

var API *softetherApi.API

var SoftHost string
var SoftPort int
var SoftPassword string

var ServerCert string
var RemoteAccess string
var DDnsHostName string
var Ipv4Address string

const OpenVpnServicePort = 21194

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
		RemoteAccess = MakeServerCipherFile(RemoteAccess, ipv4, strconv.Itoa(OpenVpnServicePort))
	}

	//
	API.CreateListener(OpenVpnServicePort, true)

	log.Println("Softether Init Success")
	CronInit()
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
func MakeServerCipherFile(config string, ipv4 string, port string) string {
	var context string
	scanner := bufio.NewScanner(strings.NewReader(config))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if string(line[0]) == "#" || string(line[0]) == ";" {
			continue
		}
		if string(line[0:6]) == "remote" {
			continue
		}
		context += fmt.Sprintf("%s\n", string(line))
	}
	return context
}
