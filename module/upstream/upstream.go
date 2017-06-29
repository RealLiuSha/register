package upstream

import (
	"fmt"
	"path"
	"regexp"
	"errors"
	"strings"

	"github.com/itchenyi/register/internal"
	"github.com/itchenyi/register/internal/log"
	"github.com/itchenyi/register/internal/grequest"
)

func generate(address string, port string) map[string]string {
	return map[string]string{
		"address": address,
		"port": port,
		"weight": "1",
		"max_fails": "3",
		"fail_timeout": "30",
	}
}

func contains(upstreamList []map[string]string, upstream map[string]string) bool {
	for _, _upstream := range upstreamList {
		if _upstream["address"] == upstream["address"] && _upstream["port"] == upstream["port"] {
			return true
		}
	}

	return false
}

func DynamicServers() ([]string, error) {
	targetUrl := fmt.Sprintf("http://%s:%s/v2/", internal.MARATHON_HOST, internal.MARATHON_PORT) +
		path.Join("apps", internal.MARATHON_DYUPS_URL)
	log.Debug("the orange api for marathon is: ", targetUrl)

	resData := make(map[string]interface{})
	request := grequest.Q()

	resp, _, respErr := request.Get(targetUrl).EndStruct(&resData)

	if respErr != nil || resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("connection marathon service failed: %s", respErr))
	}

	servers := make([]string, 0, len(resData))
	AppData, ok := resData["app"].(map[string]interface{})
	if !ok {
		return nil, errors.New("not found 'app' in response...")
	}

	for _, task := range AppData["tasks"].([]interface{}) {
		taskData, ok := task.(map[string]interface{})
		if !ok {
			return nil, errors.New("not found 'tasks' in response...")
		}

		ipAddresses, ok := taskData["ipAddresses"].([]interface{})
		if !ok {
			return nil, errors.New("not found 'ipAddresses' in response...")
		}

		for _, ipAddressMap := range ipAddresses {
			if ipAddressMap.(map[string]interface{})["protocol"].(string) != "IPv4" {
				continue
			}

			servers = append(servers, ipAddressMap.(map[string]interface{})["ipAddress"].(string))
		}
	}

	return servers, nil
}

func Servers(name string, svcAddress string) ([]map[string]string, error) {
	targetUrl := fmt.Sprintf("http://%s:%s", svcAddress, internal.DYUPS_PORT) +
		path.Join(internal.DYUPS_URL, fmt.Sprintf("?upstream=%s&verbose=true", name))

	log.Debug("the upstream server list api for openresty is: ", targetUrl)

	servers := make([]map[string]string, 0)
	request := grequest.Q()

	resp, resData, respErr := request.Get(targetUrl).EndBytes()

	if respErr != nil || resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("get upstream server list failed:%s", respErr))
	}

	for _, serverText := range strings.Split(string(resData), "\n") {
		if serverText == "" { continue }

		var serverExp = regexp.MustCompile(
			`(?P<address>\d+.\d+.\d+.\d+):(?P<port>\d+)\s` +
				`weight=(?P<weight>\d+)\s` +
				`max_fails=(?P<max_fails>\d+)\s` +
				`fail_timeout=(?P<fail_timeout>\d+)` +
				`\s?(?P<down>down)?`)

		server := make(map[string]string)
		serverMatch := serverExp.FindStringSubmatch(serverText)

		for i, subExpName := range serverExp.SubexpNames() {
			if i != 0 { server[subExpName] = serverMatch[i] }
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func Update(name string, address string, port string, action string) bool {
	dynamicServers, err := DynamicServers()
	if err != nil {
		log.Error("get dynamic upstream server list failed:", err)
		return false
	}

	for _, dynamicServer := range dynamicServers {
		servers, serversErr := Servers(name, dynamicServer)
		if serversErr != nil {
			log.Error("get upstream server list failed: ", serversErr)
			return false
		}

		if contains(servers, generate(address, port)) {
			targetUrl := fmt.Sprintf("http://%s:%s", dynamicServer, internal.DYUPS_PORT) +
				path.Join(internal.DYUPS_URL)

			request := grequest.Q().SetRawQuery(true)

			switch action {
			case "add", "remove":
				resp, _, respErr := request.Get(targetUrl).
					Query(fmt.Sprintf("upstream=%s&%s=true", name, action)).
					Query(fmt.Sprintf("server=%s:%s", address, port)).
					Query("weight=1&max_fails=3&fail_timeout=30&verbose=true").
					End()

				if respErr != nil || resp.StatusCode != 200 {
					fmt.Println(resp)
					log.Error("update upstream failed: ", respErr)
					return false
				}

			default:
				log.Error("not match action: ", action)
			}
		} else {

		}
	}

	return true
}
