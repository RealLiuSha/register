package modules

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"git.wolaidai.com/DevOps/register/pkg/cache"
	"git.wolaidai.com/DevOps/register/pkg/proto"
	"git.wolaidai.com/DevOps/register/pkg/utils"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
)

type envMap map[string]interface{}

func GetContainerEnv(container *docker.Container, err error) (envMap, error) {
	if err != nil {
		return nil, err
	}

	env := make(envMap)

	for _, envItem := range container.Config.Env {
		key, val := func(s, sep string) (string, string) {
			env_splits := strings.Split(s, sep)
			return env_splits[0], env_splits[1]
		}(envItem, "=")

		env[key] = val
	}

	if !func(_env envMap) bool {
		_keys := []string{"SERVICE_NAME", "SERVICE_PORT", "MARATHON_APP_ID"}

		for _, key := range _keys {
			if _, ok := _env[key]; !ok {
				return false
			}

			if key == "SERVICE_PORT" {
				port, err := strconv.Atoi(_env[key].(string))

				if err != nil {
					return false
				}
				_env[key] = port
			}
		}

		return true
	}(env) {
		return nil, errors.New("Keys and EnvKeys Not Match...")
	}

	env["DOCKER_ADDRESS"] = func(_networks map[string]docker.ContainerNetwork, _id string) string {
		for key := range _networks {
			if obj, exists := _networks[key]; exists {
				if obj.IPAddress != "" {
					return obj.IPAddress
				}

				continue
			}
		}

		cmd := fmt.Sprintf("docker exec %s ip addr show ", container.ID) +
			"eth0|awk '/inet /{print $2}'|cut -d/ -f1"

		address, err := utils.CmdOutBytes("/bin/sh", "-c", cmd)
		if err != nil {
			return ""
		}

		return strings.TrimSuffix(string(address), "\n")
	}(container.NetworkSettings.Networks, container.ID)

	cache.Cache.AddEnv(container.ID, env)
	return env, err
}

func EventHandle(event *docker.APIEvents, client *docker.Client) {
	getRegId := func(servicePort int) string {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		return fmt.Sprintf("%s:%s:%d", hostname, event.ID, servicePort)
	}

	switch event.Action {
	case "start", "unpause":
		log.Info("Event:start --> Try to register services")

		containerEnv, err := GetContainerEnv(client.InspectContainer(event.ID))
		if err != nil {
			log.Error(fmt.Sprintf("get envs error: %s, please check container...", err))
			return
		}

		err = svcRegister(proto.Register{
			Id:                getRegId(containerEnv["SERVICE_PORT"].(int)),
			Name:              containerEnv["SERVICE_NAME"].(string),
			Address:           containerEnv["DOCKER_ADDRESS"].(string),
			Tags:              []string{containerEnv["MARATHON_APP_ID"].(string)},
			Port:              containerEnv["SERVICE_PORT"].(int),
			EnableTagOverride: false,
		})

		if err != nil {
			log.Error(err)
			return
		}

	case "pause", "die":
		log.Info("Event:stop --> Try to Unregister services")
		containerEnv, err := cache.Cache.GetEnv(event.ID)
		if err != nil {
			log.Error(fmt.Sprintf("get envs error: %s, please check container...", err))
			return
		}

		if err := svcDeRegister(proto.Register{
			Id:                getRegId(containerEnv["SERVICE_PORT"].(int)),
			Name:              containerEnv["SERVICE_NAME"].(string),
			Address:           containerEnv["DOCKER_ADDRESS"].(string),
			Tags:              []string{containerEnv["MARATHON_APP_ID"].(string)},
			Port:              containerEnv["SERVICE_PORT"].(int),
			EnableTagOverride: false,
		}); err == nil { cache.Cache.DelEnv(event.ID) }
	}
}
