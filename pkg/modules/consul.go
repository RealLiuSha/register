package modules

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"git.wolaidai.com/DevOps/register/pkg/g"
	"git.wolaidai.com/DevOps/register/pkg/utils"
	"git.wolaidai.com/DevOps/register/pkg/utils/grequest"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
)

func ConsulEndpoint() string {
	var address, network string
	if utils.IpValid(g.CONSUL_HOST) {
		return fmt.Sprintf("http://%s:%s/v1/", g.CONSUL_HOST, g.CONSUL_PORT)
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

	if !utils.IpValid(address) {
		log.Error("Error: Failed to get available address")
		os.Exit(0)
	}

	return fmt.Sprintf("http://%s:%s/v1/", address, g.CONSUL_PORT)
}

func ConsulKvUpstreams() ([]string, error) {
	targetUrl := ConsulEndpoint() + path.Join("kv", "paas", "ngx", "upstream_name")
	log.Debug("the upstream list api for consul-kv is: ", targetUrl)

	resData := make([]interface{}, 1)
	request := grequest.Q()

	log.Debug("Consule Server:", targetUrl)
	resp, _, resp_err := request.Get(targetUrl).EndStruct(&resData)

	if resp_err != nil || resp.StatusCode != 200 {
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
