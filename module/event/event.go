package event

import (
	"os"
	"fmt"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"github.com/itchenyi/register/proto"
	"github.com/itchenyi/register/internal/log"
	"github.com/itchenyi/register/module/service"
)

func Handle (event *docker.APIEvents, client *docker.Client) {
	type envMap map[string]interface{}

	getContainerEnv := func (container *docker.Container, err error) (envMap, error) {
		if err != nil {
			return nil, err
		}

		envs := make(envMap)

		for _, env := range container.Config.Env {
			key, val := func (s, sep string) (string, string) {
				env_splits := strings.Split(s, sep)
				return env_splits[0], env_splits[1]
			}(env, "=")

			envs[key] = val
		}

		if !func(_envs envMap) bool {
			_keys := []string{"HOST", "SERVICE_NAME", "SERVICE_PORT", "SERVICE_TAGS"}

			for _, key := range _keys {
				if _, ok := _envs[key]; !ok {
					return false
				}

				if key == "SERVICE_PORT" {
					port, err := strconv.Atoi(_envs[key].(string))

					if err != nil {
						return false
					}
					_envs[key] = port
				}
			}

			return true
		}(envs) { return nil, errors.New("Keys and EnvKeys Not Match...")}

		envs["DOCKER_ADDRESS"] = func (_networks map[string]docker.ContainerNetwork, _id string) string {
			for key := range _networks {
				if obj, exists := _networks[key]; exists {
					return obj.IPAddress
				}
			}

			return func () string {
				cmd := exec.Command(
					fmt.Sprintf("docker exec %s ifconfig ", _id) +
						"eth0|grep -oP '\\d.+(?=  (Bcast:|netmask))'")

				out, _ := cmd.CombinedOutput()

				return string(out)
			} ()
		} (container.NetworkSettings.Networks, container.ID)

		return envs, err
	}

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

		containerEnv, err := getContainerEnv(client.InspectContainer(event.ID))
		if err != nil {
			log.Error(fmt.Sprintf("get envs error: %s, please check container...", err))
			return
		}

		err = service.Register(proto.Register {
			Id: getRegId(containerEnv["SERVICE_PORT"].(int)),
			Name: containerEnv["SERVICE_NAME"].(string),
			Address: containerEnv["HOST"].(string),
			Tags: []string{containerEnv["SERVICE_TAGS"].(string)},
			Port: containerEnv["SERVICE_PORT"].(int),
			EnableTagOverride: false,
		}, containerEnv["DOCKER_ADDRESS"].(string))

		if err !=  nil {
			log.Error(err)
			return
		}

	case "pause", "die":
		log.Info("Event:stop --> Try to Unregister services")
		//result, _ := json.Marshal(event)
		//log.Info(string(result))
		containerEnv, err := getContainerEnv(client.InspectContainer(event.ID))
		if err != nil {
			log.Error(fmt.Sprintf("get envs error: %s, please check container...", err))
			return
		}

		service.DeRegister(proto.Register {
			Id: getRegId(containerEnv["SERVICE_PORT"].(int)),
			Name: containerEnv["SERVICE_NAME"].(string),
			Address: containerEnv["HOST"].(string),
			Tags: []string{containerEnv["SERVICE_TAGS"].(string)},
			Port: containerEnv["SERVICE_PORT"].(int),
			EnableTagOverride: false,
		})
	}
}

