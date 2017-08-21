package util

import (
	"net"
	"strings"
	"fmt"
)

const REDIS_ALIVE_KEY = "diconf_alive"
const REDIS_MESSAGE_KEY = "disconf_message"
const REDIS_RESULT_KEY = "disconf_result_%s"

type Message struct {
	Sid  string `json:"sid"`
	Dest string `json:"dest"`
	Data string        `json:"data"`
}

func (msg Message) ToString() string {
	return fmt.Sprintf("{Sid:%s,Dest:%s,Data:%s}", msg.Sid, msg.Dest, msg.Data)
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		arr := strings.Split(addr.String(), "/")
		if strings.HasPrefix(arr[0], "192.168") || strings.HasPrefix(arr[0], "172.10") {
			return arr[0], nil
		}
	}
	return "", nil
}