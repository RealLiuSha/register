package main

import (
	"os"
	"fmt"
	"flag"
	"github.com/itchenyi/register/internal"
	"github.com/itchenyi/register/internal/log"
	"github.com/itchenyi/register/internal/gpool"
	"github.com/itchenyi/register/module/service"
	"github.com/itchenyi/register/module/event"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Consul Service Registration Program with enjoy!\n\n")
		flag.PrintDefaults()
	}

	flag.IntVar(&internal.CONCURRENCY, "concurrency", 10, "concurrency number")
	flag.StringVar(&internal.DOCKER_ENDPOINT, "docker.endpoint", "unix:///var/run/docker.sock", "Docker Conn EndPoint")

	flag.StringVar(&internal.CONSUL_HOST, "consul.host", "", "consul server host")
	flag.StringVar(&internal.CONSUL_PORT, "consul.port", "8500", "consul server port")

	flag.StringVar(&internal.MARATHON_HOST, "marathon.host", "192.168.20.2", "marathon server host")
	flag.StringVar(&internal.MARATHON_PORT, "marathon.port", "8080", "marathon server port")
	flag.StringVar(&internal.MARATHON_DYUPS_URL, "marathon.dyups", "/base-service/loadbalance/orange", "marathon of orange api url")

	flag.StringVar(&internal.DYUPS_PORT, "dyups.port", "18081", "orange server port")
	flag.StringVar(&internal.DYUPS_URL, "dyups.url", "/upstream-admin", "orange upstream url")

	flag.Int64Var(&internal.UPSTREAM_CHECK_SECONDS, "check.second", 10, "timer check upstream duration")
	flag.BoolVar(&internal.UPSTREAM_CHECK_ENABLE, "check.enable", false, "timer check upstream is enable")
	flag.Parse()
}

func main() {
	if internal.UPSTREAM_CHECK_ENABLE {
		go service.CheckTimer(internal.UPSTREAM_CHECK_SECONDS)
	}

	client, err := docker.NewClient(internal.DOCKER_ENDPOINT)
	if err != nil {
		log.Error("get docker client instance failure ", err)
		os.Exit(1)
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

	p := gpool.NewPool(internal.CONCURRENCY * 2, internal.CONCURRENCY)
	defer p.Release()

	for {
		select {
		case e, ok := <-listener:
			if !ok {
				log.Error(fmt.Sprintf("'%s' not found or permission denied...", internal.DOCKER_ENDPOINT))
				return
			}
			p.JobQueue <- func () {
				event.Handle(e, client)
			}
		}
	}
}
