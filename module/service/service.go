package service

import (
	"fmt"
	"time"
	"path"
	"errors"
	"strings"
	"strconv"
	"encoding/json"

	"github.com/itchenyi/register/proto"
	"github.com/itchenyi/register/internal/log"
	"github.com/itchenyi/register/module/consul"
	"github.com/itchenyi/register/internal/tool"
	"github.com/itchenyi/register/module/upstream"
	"github.com/itchenyi/register/internal/grequest"
)

func isRegister(serviceId string) bool {
	resData := make(map[string]interface{})

	_serivces := func () []string {
		request := grequest.New().Timeout(5*time.Second)
		targetUrl := consul.ServerURL() + path.Join("agent", "services")

		resp, _, respErr := request.Put(targetUrl).EndStruct(&resData)

		if respErr != nil || resp.StatusCode != 200 {
			return []string{}
		}

		respKeys := make([]string, 0, len(resData))
		for key := range resData {
			respKeys = append(respKeys, key)
		}

		return respKeys
	}()

	for _, _service := range _serivces {
		if _service == serviceId {
			return true
		}
	}

	return false
}

func Register (register proto.Register, svcAddress string) error {
	for retry := 0; retry < 30; retry++ {
		time.Sleep(time.Second * 2)

		if tool.HttpHealthCheck(svcAddress, strconv.Itoa(register.Port)) {
			break
		}

		if retry < 15 {
			log.Info("try to verify http server: count-->:", retry + 1)
			continue
		}

		log.Error("try to verify http server: count-->:", retry + 1)
		if (retry + 1) == 30 {
			return errors.New("service connection fails...")
		}
	}

	request := grequest.New().Timeout(5*time.Second)
	targetUrl := consul.ServerURL() + path.Join("agent", "service", "register")
	log.Info("consul server uri: ", targetUrl)

	data, err := json.Marshal(register)
	if err != nil {
		return errors.New(fmt.Sprintf("service request data error: %s", err))
	}

	resp, _, respErr := request.Put(targetUrl).Send(string(data)).End()
	if respErr != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service register failure: %s", respErr))
	}

	for retry := 0; retry < 5; retry++ {
		time.Sleep(time.Second)
		if isRegister(register.Id) {
			break
		}

		log.Debug("detect register status: count-->:", retry + 1)
	}

	log.Info("service register successful: ID-->", register.Id)

	healthCheck := proto.HealthCheck{
		Interval: "10s",
		Timeout: "3s",
		Deregister_Csa: "1m",
		ServiceID: register.Id,
	}
	id_splits := strings.Split(register.Id, ":")

	func () {
		healthCheck.Name = fmt.Sprintf("ping_check@%s", id_splits[1][0:12])
		healthCheck.Script = "/welab.co/bin/consul-health ping"
		healthCheck.Tcp = ""

		request := grequest.New().Timeout(5*time.Second)
		targetUrl := consul.ServerURL() + path.Join("agent", "check", "register")

		data, err := json.Marshal(healthCheck)

		if err != nil {
			log.Error("service request data error: ", err)
		}

		resp, _, respErr := request.Put(targetUrl).Send(string(data)).End()

		if respErr != nil || resp.StatusCode != 200 {
			log.Error("service health check(ping_check) register failure: ", respErr)
		}

		log.Info("service health check(ping_check) register successful")
	}()

	func () {
		healthCheck.Name = fmt.Sprintf("tcp_check@%s", id_splits[1][0:12])
		healthCheck.Tcp = fmt.Sprintf("%s:%s", svcAddress, strconv.Itoa(register.Port))
		healthCheck.Script = ""

		request := grequest.New().Timeout(5*time.Second)
		targetUrl := consul.ServerURL() + path.Join("agent", "check", "register")

		data, err := json.Marshal(healthCheck)

		if err != nil {
			log.Error("service request data error: ", err)
		}

		resp, _, respErr := request.Put(targetUrl).Send(string(data)).End()

		if respErr != nil || resp.StatusCode != 200 {
			log.Error("service health check(tcp_check) register failure: ", respErr)
		}

		log.Info("service health check(tcp_check) register successful")
	}()

	func () {
		upstreamNameList, err := consul.UpstreamList()
		if err != nil {
			log.Error("update upstream failure: ", err)
		}

		for _, upstreamNmae := range upstreamNameList {
			if upstreamNmae == register.Name {
				log.Debug("start updating upstream(%s)...")
				upstream.Update(upstreamNmae, register.Address, strconv.Itoa(register.Port), "add")
			}
		}

		log.Error(fmt.Sprintf("%s not in upstream list...", register.Name))
	}()

	return nil
}

func DeRegister (register proto.Register) error {
	request := grequest.New().Timeout(5*time.Second)
	targetUrl := consul.ServerURL() + path.Join(
		"agent", "service", "deregister", register.Id)

	log.Debug("consul server uri: ", targetUrl)

	resp, _, respErr := request.Put(targetUrl).End()

	if respErr != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service deregister failure: %s", respErr))
	}

	for retry := 0; retry == 5; retry++ {
		time.Sleep(time.Second)
		if !isRegister(register.Id) {
			break
		}

		log.Debug("detect deregister status: count->>", retry + 1)
	}

	log.Info("service deregister successful: ID-->", register.Id)

	func () {
		upstreamNameList, err := consul.UpstreamList()
		if err != nil {
			log.Error("update upstream failure: ", err)
		}

		for _, upstreamName := range upstreamNameList {
			if upstreamName == register.Name {
				log.Debug("start updating upstream(%s)...")
				upstream.Update(upstreamName, register.Address, strconv.Itoa(register.Port), "add")
			}
		}

		log.Error(fmt.Sprintf("%s not in upstream list...", register.Name))
	}()

	return nil
}

func CheckTimer (seconds int64) {
	timer := time.NewTicker(time.Duration(seconds) * time.Second)
	for {
		select {
		case <- timer.C:
			// TODO Upstream Integrity check
			log.Info("Start Upstream Check...")
		}
	}
	//timer.Stop()
}
