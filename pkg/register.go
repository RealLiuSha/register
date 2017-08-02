package pkg

import (
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/itchenyi/gpool"

	"git.wolaidai.com/DevOps/register/pkg/g"
	"git.wolaidai.com/DevOps/register/pkg/modules"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
)

func Start() {
	if g.UPSTREAM_CHECK_ENABLE {
		go modules.SvcCheckTimer(g.UPSTREAM_CHECK_SECONDS)
	}

	client, err := docker.NewClient(g.DOCKER_ENDPOINT)
	if err != nil {
		log.Error("failure to get docker client instance:", err)
		os.Exit(1)
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Error("failure to get containers:", err)
		os.Exit(1)
	}

	for _, container := range containers {
		_, err := modules.GetContainerEnv(client.InspectContainer(container.ID))
		if err != nil {
			log.Error(fmt.Sprintf("get envs error: %s, please check container...", err))
			continue
		}
	}

	listener := make(chan *docker.APIEvents)
	err = client.AddEventListener(listener)
	if err != nil {
		log.Error("add event listener failure ", err)
		os.Exit(1)
	}

	defer func() {
		err = client.RemoveEventListener(listener)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}()

	p := gpool.NewPool(g.CONCURRENCY*2, g.CONCURRENCY)
	defer p.Release()

	for {
		select {
		case e, ok := <-listener:
			if !ok {
				log.Error(fmt.Sprintf("'%s' not found or permission denied...", g.DOCKER_ENDPOINT))
				return
			}
			p.JobQueue <- func() {
				modules.EventHandle(e, client)
			}
		}
	}
}
