package tool

import (
	"fmt"
	"net"

	"github.com/itchenyi/register/internal/log"
)

func IpValid(address string) bool {
	trial := net.ParseIP(address)
	if trial.To4() == nil {
		return false
	}
	return true
}

func HttpHealthCheck(host string, port string) bool {
	servAddr := fmt.Sprintf("%s:%s", host, port)

	echoPacket := "GET / HTTP/1.1\r\n" +
		fmt.Sprintf("Host: %s\r\n", servAddr) +
		"Accept: */*\r\n" +
		"User-Agent: DevOps/Regsiter Check\r\n\r\n"

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		log.Error(fmt.Sprintf("ResolveTCPAddr(%s) failed:", servAddr), err.Error())
		return false
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Error(fmt.Sprintf("Dial(%s) failed:", servAddr), err.Error())
		return false
	}

	defer conn.Close()

	_, err = conn.Write([]byte(echoPacket))
	if err != nil {
		log.Info(fmt.Sprintf("Write to server(%s) failed:", servAddr), err.Error())
		return false
	}

	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		log.Info(fmt.Sprintf("Write to server(%s) failed:", servAddr), err.Error())
		return false
	}

	return true
}
