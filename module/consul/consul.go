package consul

import (
	"os"
	"fmt"
	"net"
	"path"
	"errors"
	"strings"
	"encoding/base64"

	"github.com/itchenyi/register/internal"
	"github.com/itchenyi/register/internal/log"
	"github.com/itchenyi/register/internal/tool"
	"github.com/itchenyi/register/internal/grequest"
)

func ServerURL() string {
	var address, network string
	if tool.IpValid(internal.CONSUL_HOST) {
		return fmt.Sprintf("http://%s:%s/v1/", internal.CONSUL_HOST, internal.CONSUL_PORT)
	}

	network = "192.168.0.0/16"
	_, networkSubnet, _ := net.ParseCIDR(network)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error("generate consul server url failure: ", err)
		os.Exit(1)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if (ipnet.IP.To4() != nil) && networkSubnet.Contains(ipnet.IP.To4()) {
				address = ipnet.IP.String()
				break
			}
		}
	}

	if !tool.IpValid(address) {
		log.Error("Error: Failed to get available address")
		os.Exit(0)
	}

	return fmt.Sprintf("http://%s:%s/v1/", address, internal.CONSUL_PORT)
}


func UpstreamList() ([]string, error) {
	targetUrl := ServerURL() + path.Join("kv", "paas", "ngx", "upstream_name")
	log.Debug("the upstream list api for consul-kv is: ", targetUrl)

	resData := make([]interface{}, 1)
	request := grequest.Q()

	resp, _, resp_err := request.Get(targetUrl).EndStruct(&resData)

	if resp_err != nil || resp.StatusCode != 200 {
		fmt.Println(resp)
		return nil, errors.New(fmt.Sprintf("connection consul service failed: %s", resp_err))
	}

	resMap, ok := resData[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("not found 'data' in response...")
	}

	valueString, err := base64.StdEncoding.DecodeString(resMap["Value"].(string))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("get upstream list failed: %s", err))
	}

	return strings.Split(string(valueString), ","), nil
}
