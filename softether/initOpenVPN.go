package softether

import (
	"bufio"
	"fmt"
	"github.com/kiddnoke/SoftetherGo"
	"log"
	"proxy-golang/common"
	"strconv"
	"strings"
)

var SoftHost string
var SoftPort int
var SoftPassword string

var ServerCert string
var RemoteAccess string
var DDnsHostName string
var Ipv4Address string

const OpenVpnServicePort = 443

func Init() {
	selflogger = common.NewLogger(common.LOG_DEFAULT, "CronTask")

	if checkSoftetherIsFirst() == true {
		if err := softetherFirstInit(SoftPassword); err != nil {
			panic(err)
		}
	} else {
		PoolConnect()
		_, err := PoolGetConn()
		if err != nil {
			panic(err)
		}
	}

	//
	c, _ := PoolGetConn()

	if remoteaccessfile, err := c.GetOpenVpnRemoteAccess(); err != nil {
		panic(err)
	} else {
		RemoteAccess = remoteaccessfile
	}
	//
	if cert, err := c.GetServerCert(); err != nil {
		panic(err)
	} else {
		ServerCert = cert
	}
	//
	if hostname, ipv4, err := c.GetDDnsHostName(); err != nil {
		panic(err)
	} else {
		DDnsHostName = hostname
		Ipv4Address = ipv4
		RemoteAccess = MakeServerCipherFile(RemoteAccess, ipv4, strconv.Itoa(OpenVpnServicePort))
	}

	//
	c.CreateListener(OpenVpnServicePort, true)

	log.Println("Softether Init Success")
	PoolHeartBeatLoop()
	CronInit()
}

func checkSoftetherIsFirst() bool {
	checkSoftetherIsFirstApi := softetherApi.NewAPI(SoftHost, SoftPort, "")
	if err := checkSoftetherIsFirstApi.HandShake(); err != nil {
		checkSoftetherIsFirstApi = nil
		return false
	} else {
		defer checkSoftetherIsFirstApi.Disconnect()
		return true
	}

}
func softetherFirstInit(password string) error {
	Api := softetherApi.NewAPI(SoftHost, SoftPort, "")
	defer Api.Disconnect()
	if err := Api.HandShake(); err != nil {
		if _, err := Api.SetServerPassword(password); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return err
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
