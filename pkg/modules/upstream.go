package modules

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"git.wolaidai.com/DevOps/register/pkg/g"
	"git.wolaidai.com/DevOps/register/pkg/utils/grequest"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
)

func upstreamGenerate(address string, port string) map[string]string {
	return map[string]string{
		"address":      address,
		"port":         port,
		"weight":       "1",
		"max_fails":    "3",
		"fail_timeout": "30",
	}
}

func upstreamContains(upstreamList []map[string]string, upstream map[string]string) bool {
	log.Debug("upstream:", upstream)
	log.Debug("upstreamList:", upstreamList)
	for _, _upstream := range upstreamList {
		if _upstream["address"] == upstream["address"] && _upstream["port"] == upstream["port"] {
			return true
		}
	}

	return false
}

func dyUpstreamServers() ([]string, error) {
	targetUrl := fmt.Sprintf("http://%s:%s/v2/", g.MARATHON_HOST, g.MARATHON_PORT) +
		path.Join("apps", g.MARATHON_DYUPS_URL)
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

func upstreamServers(name string, svcAddress string) ([]map[string]string, error) {
	targetUrl := fmt.Sprintf("http://%s:%s", svcAddress, g.DYUPS_PORT) +
		path.Join(g.DYUPS_URL, fmt.Sprintf("?upstream=%s&verbose=true", name))

	log.Debug("the upstream server list api for openresty is: ", targetUrl)

	servers := make([]map[string]string, 0)
	request := grequest.Q()

	resp, resData, respErr := request.Get(targetUrl).EndBytes()

	if respErr != nil || resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("get upstream server list failed:%s", respErr))
	}

	for _, serverText := range strings.Split(string(resData), "\n") {
		if serverText == "" {
			continue
		}

		var serverExp = regexp.MustCompile(
			`(?P<address>\d+.\d+.\d+.\d+):(?P<port>\d+)\s` +
				`weight=(?P<weight>\d+)\s` +
				`max_fails=(?P<max_fails>\d+)\s` +
				`fail_timeout=(?P<fail_timeout>\d+)` +
				`\s?(?P<down>down)?`)

		server := make(map[string]string)
		serverMatch := serverExp.FindStringSubmatch(serverText)

		for i, subExpName := range serverExp.SubexpNames() {
			if i != 0 {
				server[subExpName] = serverMatch[i]
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func upstreamUpdate(name string, address string, port string, action string) error {
	dynamicServers, err := dyUpstreamServers()
	if err != nil {
		return fmt.Errorf("get dynamic upstream server list failed: %s", err)
	}

	log.Debug("dynamicServers:", dynamicServers)
	log.Debug(action, ":", address, ":", port)

	for _, dynamicServer := range dynamicServers {
		servers, err := upstreamServers(name, dynamicServer)
		if err != nil {
			return fmt.Errorf("get upstream server list failed: %s", err)
		}

		request := grequest.Q().SetRawQuery(true)
		targetUrl := fmt.Sprintf("http://%s:%s", dynamicServer, g.DYUPS_PORT) +
			path.Join(g.DYUPS_URL)

		if action == "remove" && upstreamContains(servers, upstreamGenerate(address, port)) {
			if len(servers) == 1 {
				resp, respData, respErr := request.Get(targetUrl).
					Query(fmt.Sprintf("upstream=%s&%s=true", name, action)).
					Query("server=127.0.0.1:8080&weight=0&verbose=true").
					End()

				log.Debug("failed to add default upstream resp:", resp)
				log.Debug("failed to add default upstream resp data:", respData)
				if respErr != nil || resp.StatusCode != 200 {
					fmt.Errorf("failed to add defulat upstream failed: %s", respErr)
				}
			}
		}

		log.Debug("DYUpstreamServer:", targetUrl)
		resp, respData, respErr := request.Get(targetUrl).
			Query(fmt.Sprintf("upstream=%s&%s=true", name, action)).
			Query(fmt.Sprintf("server=%s:%s", address, port)).
			Query("weight=1&max_fails=3&fail_timeout=30&verbose=true").
			End()

		log.Debug("update upstream resp:", resp)
		log.Debug("update upstream resp data:", respData)

		if respErr != nil || resp.StatusCode != 200 {
			fmt.Errorf("update upstream failed: %s", respErr)
		}
	}

	return nil
}
