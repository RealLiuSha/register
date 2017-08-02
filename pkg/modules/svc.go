package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"git.wolaidai.com/DevOps/register/pkg/proto"
	"git.wolaidai.com/DevOps/register/pkg/utils"
	"git.wolaidai.com/DevOps/register/pkg/utils/grequest"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
)

func svcIsRegister(serviceId string) bool {
	resData := make(map[string]interface{})

	_serivces := func() []string {
		request := grequest.Q()
		targetUrl := ConsulEndpoint() + path.Join("agent", "services")

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

func svcRegister(register proto.Register) error {
	if !utils.IpValid(register.Address) {
		return errors.New("invalid docker IP address")
	}

	for retry := 0; retry < 90; retry++ {
		time.Sleep(time.Second * 2)

		if utils.TcpHealthCheck(register.Address, strconv.Itoa(register.Port)) {
			break
		}

		if retry < 15 {
			log.Info("try to verify server: count-->:", retry+1)
			continue
		}

		log.Error("try to verify server: count-->:", retry+1)
		if (retry + 1) == 90 {
			return errors.New("service connection fails...")
		}
	}

	request := grequest.Q()

	targetUrl := ConsulEndpoint() + path.Join("agent", "service", "register")
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
		if svcIsRegister(register.Id) {
			break
		}

		time.Sleep(time.Second)
		log.Debug("detect register status: count-->:", retry+1)
	}

	log.Info("service register successful: ID-->", register.Id)

	healthCheck := proto.HealthCheck{
		Interval:       "10s",
		Timeout:        "3s",
		Deregister_Csa: "1m",
		ServiceID:      register.Id,
	}
	id_splits := strings.Split(register.Id, ":")

	func() {
		healthCheck.Name = fmt.Sprintf("ping_check@%s", id_splits[1][0:12])
		healthCheck.Script = "/welab.co/bin/consul-health ping"
		healthCheck.Tcp = ""

		request := grequest.Q()

		targetUrl := ConsulEndpoint() + path.Join("agent", "check", "register")

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

	func() {
		healthCheck.Name = fmt.Sprintf("tcp_check@%s", id_splits[1][0:12])
		healthCheck.Tcp = fmt.Sprintf("%s:%s", register.Address, strconv.Itoa(register.Port))
		healthCheck.Script = ""

		request := grequest.Q()

		targetUrl := ConsulEndpoint() + path.Join("agent", "check", "register")

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

	upstreamNameList, err := ConsulKvUpstreams()
	if err != nil {
		log.Error("update upstream failure: ", err)
	}

	for _, upstreamNmae := range upstreamNameList {
		if upstreamNmae == register.Name {
			log.Debug("start update upstream:", register.Name)
			err := upstreamUpdate(register.Name, register.Address, strconv.Itoa(register.Port), "add")
			if err != nil { return err }
			return nil
		}
	}

	return fmt.Errorf("%s not in upstream list...", register.Name)
}

func svcDeRegister(register proto.Register) error {
	request := grequest.Q()

	targetUrl := ConsulEndpoint() + path.Join(
		"agent", "service", "deregister", register.Id)

	log.Debug("consul server uri: ", targetUrl)

	resp, _, respErr := request.Put(targetUrl).End()

	if respErr != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service deregister failure: %s", respErr))
	}

	for retry := 0; retry == 5; retry++ {
		time.Sleep(time.Second)
		if !svcIsRegister(register.Id) {
			break
		}

		log.Debug("detect deregister status: count->>", retry+1)
	}

	log.Info("service deregister successful: ID-->", register.Id)

	upstreamNameList, err := ConsulKvUpstreams()
	if err != nil {
		log.Error("update upstream failure: ", err)
	}

	for _, upstreamName := range upstreamNameList {
		if upstreamName == register.Name {
			log.Debug("start remove upstream:", register.Name)
			err := upstreamUpdate(upstreamName, register.Address, strconv.Itoa(register.Port), "remove")
			if err != nil { return err }
			return nil
		}
	}

	return fmt.Errorf("%s not in upstream list...", register.Name)
}

func SvcCheckTimer(seconds int64) {
	timer := time.NewTicker(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-timer.C:
			// TODO Upstream Integrity check
			log.Info("Start Upstream Check...")
		}
	}
	//timer.Stop()
}
